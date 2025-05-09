package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/DIMO-Network/identity-api/graph"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/kafka"
	"github.com/DIMO-Network/shared/pkg/settings"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "identity-api").Logger()

	settings, err := settings.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't load settings.")
	}

	zl, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Error parsing log level from %q.", settings.LogLevel)
	}

	logger = logger.Level(zl)

	logger.Info().Msgf("Loaded configuration. Addresses: Registry %s, Vehicle %s, Aftermarket Device %s.", settings.DIMORegistryAddr, settings.VehicleNFTAddr, settings.AftermarketDeviceAddr)

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		command := "up"
		if len(os.Args) > 2 {
			command = os.Args[2]
			if command == "down-to" || command == "up-to" {
				command = command + " " + os.Args[3]
			}
		}
		migrateDatabase(logger, &settings, command)
		return
	}

	dbs := db.NewDbConnectionFromSettings(context.Background(), &settings.DB, true)
	dbs.WaitForDB(logger)

	startContractEventsConsumer(ctx, &logger, &settings, dbs)

	repoLogger := logger.With().Str("component", "repository").Logger()
	baseRepo := base.NewRepository(dbs, settings, &repoLogger)

	cfg := graph.Config{Resolvers: graph.NewResolver(baseRepo)}

	serveMonitoring(strconv.Itoa(settings.MonPort), &logger)

	s := newDefaultServer(graph.NewExecutableSchema(cfg))

	srv := loader.Middleware(dbs, s, settings)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	logger.Info().Msgf("Server started on port: %d", settings.Port)

	logger.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil)).Msg("Server shut down.")
}

func startContractEventsConsumer(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, dbs db.Store) {
	kc := kafka.Config{
		Brokers: strings.Split(settings.KafkaBrokers, ","),
		Topic:   settings.ContractsEventTopic,
		Group:   "identity-api",
	}

	cevConsumer := services.NewContractsEventsConsumer(dbs, logger, settings)

	if err := kafka.Consume(ctx, kc, cevConsumer.Process, logger); err != nil {
		logger.Fatal().Err(err).Msg("Couldn't start event consumer.")
	}

	logger.Info().Msg("Contract events consumer started.")
}

func serveMonitoring(port string, logger *zerolog.Logger) *fiber.App {
	logger.Info().Str("port", port).Msg("Starting monitoring web server.")

	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	monApp.Get("/", func(c *fiber.Ctx) error { return nil })
	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	go func() {
		if err := monApp.Listen(":" + port); err != nil {
			logger.Fatal().Err(err).Str("port", port).Msg("Failed to start monitoring web server.")
		}
	}()

	return monApp
}

func newDefaultServer(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return srv
}

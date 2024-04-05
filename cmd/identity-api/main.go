package main

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/DIMO-Network/identity-api/graph"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
)

func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "identity-api").Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't load settings.")
	}

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
	cfg.Directives.OneOf = func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		// The directive on its own is advisory; everything is enforced inside of the resolver
		return next(ctx)
	}

	serveMonitoring(strconv.Itoa(settings.MonPort), &logger)

	s := handler.NewDefaultServer(graph.NewExecutableSchema(cfg))

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

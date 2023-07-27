package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/DIMO-Network/identity-api/graph"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/kafka"
	"github.com/rs/zerolog"
)

func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "identity-api").Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't load settings.")
	}

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

	repo := repositories.NewRepository(dbs, 0)

	s := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		Repo: repo,
	}}))

	srv := loader.Middleware(dbs, s)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	logger.Info().Msgf("Server started on port: %d", settings.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil); err != nil {
		logger.Fatal().Err(err).Msg("Listener terminated.")
	}

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

	logger.Info().Msg("Contracts events consumer started")
}

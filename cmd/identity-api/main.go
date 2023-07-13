package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/DIMO-Network/identity-api/graph"
	"github.com/DIMO-Network/identity-api/internal/config"
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

	pdb := db.NewDbConnectionFromSettings(context.Background(), &settings.DB, true)
	pdb.WaitForDB(logger)

	startContractEventsConsumer(ctx, &logger, &settings, pdb)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB: pdb,
	}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	logger.Info().Msg(fmt.Sprintf("Server started on port:%d", settings.Port))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil))
}

func startContractEventsConsumer(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb db.Store) {
	kc := kafka.Config{
		Brokers: strings.Split(settings.KafkaBrokers, ","),
		Topic:   settings.ContractsEventTopic,
		Group:   "identity-api",
	}

	cevConsumer := services.NewContractsEventsConsumer(pdb, logger, settings)

	if err := kafka.Consume(ctx, kc, cevConsumer.Process, logger); err != nil {
		logger.Fatal().Err(err).Msg("Couldn't start event consumer.")
	}

	logger.Info().Msg("Contracts events consumer started")
}

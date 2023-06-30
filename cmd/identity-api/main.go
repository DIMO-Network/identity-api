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
	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	dgrpc "github.com/DIMO-Network/devices-api/pkg/grpc"

	"github.com/DIMO-Network/identity-api/graph"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/kafka"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	dSvc, conn, err := getDeviceApiGrpcClient(settings)
	if err != nil {
		log.Fatal("Error occurred initializing device definitions service")
	}
	defer conn.Close()

	ddFSvc, ddConn, err := getDeviceDefsGrpcClient(settings)
	if err != nil {
		log.Fatal("Error occurred initializing device definitions service")
	}
	defer ddConn.Close()

	startContractEventsConsumer(ctx, logger, &settings, pdb, dSvc, ddFSvc)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		Db: pdb,
	}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	logger.Info().Msg(fmt.Sprintf("Server started on port:%d", settings.Port))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil))
}

func startContractEventsConsumer(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb db.Store, dSvc dgrpc.UserDeviceServiceClient, ddFSvc ddgrpc.DeviceDefinitionServiceClient) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.ContractsEventTopic,
		GroupID:         "user-devices",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start contract event consumer")
	}

	cevConsumer := services.NewContractsEventsConsumer(ctx, pdb, &logger, settings, dSvc, ddFSvc)
	consumer.Start(context.Background(), cevConsumer.ProcessContractsEventsMessages)

	logger.Info().Msg("Contracts events consumer started")
}

// getDeviceDefsGrpcClient instanties new connection with client to dd service. You must defer conn.close from returned connection
func getDeviceDefsGrpcClient(settings config.Settings) (ddgrpc.DeviceDefinitionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(settings.DefinitionsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	definitionsClient := ddgrpc.NewDeviceDefinitionServiceClient(conn)
	return definitionsClient, conn, nil
}

// getDeviceApiGrpcClient instanties new connection with client to devices-api service. You must defer conn.close from returned connection
func getDeviceApiGrpcClient(settings config.Settings) (dgrpc.UserDeviceServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(settings.DevicesApiGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, conn, err
	}
	devicesApiClient := dgrpc.NewUserDeviceServiceClient(conn)
	return devicesApiClient, conn, nil
}

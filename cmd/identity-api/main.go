package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/DIMO-Network/identity-api/graph"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

func main() {
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

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB: pdb,
	}}))

	v, e := models.Vehicles().All(context.Background(), pdb.DBS().Reader)
	if e != nil {
		fmt.Println("cant read db: ", e)
	}

	for _, r := range v {
		ad := models.AftermarketDevice{
			ID:                 r.ID,
			OwnerAddress:       r.OwnerAddress,
			BeneficiaryAddress: r.OwnerAddress,
			VehicleID:          types.NullDecimal(r.ID),
			MintTime:           time.Now(),
		}
		err = ad.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
		if err != nil {
			fmt.Println("failed to insert into db ", err)
		}

		fmt.Println(common.BytesToAddress(r.OwnerAddress.Bytes))
	}

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	prt := 4000
	logger.Info().Msg(fmt.Sprintf("Server started on port:%d", prt))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", prt), nil))
}

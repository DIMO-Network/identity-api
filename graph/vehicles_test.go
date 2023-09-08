package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

func TestVehiclesThing(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	repo := repositories.New(pdb)
	resolver := NewResolver(repo)

	logger := zerolog.New(os.Stdout)
	settings := config.Settings{
		VehicleNFTAddr:      "0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF",
		DIMORegistryChainID: 137,
	}

	contractEventConsumer := services.NewContractsEventsConsumer(pdb, &logger, &settings)

	err := contractEventConsumer.Process(ctx, &shared.CloudEvent[json.RawMessage]{
		Source: "chain/137",
		Type:   "zone.dimo.contract.event",
		Data: json.RawMessage(`
		{
			"contract": "0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF",
			"eventName": "Transfer",
			"arguments": {
				"from": "0x0000000000000000000000000000000000000000",
				"to": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
				"tokenId": 1
			}
		}
	`)})
	if err != nil {
		t.Fatal(err)
	}

	c := client.New(loader.Middleware(pdb, handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))))

	type vehicleResponse struct {
		AccessibleVehicles struct {
			Edges []struct {
				Node struct {
					ID string
				}
			}
		}
	}

	var vr vehicleResponse

	c.MustPost(`
		query OwnerVehicles($owner: Address!) {
			accessibleVehicles(address: $owner) {
				edges {
					node {
						id
					}
				}
			}
		}
	`, &vr, client.Var("owner", "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"))

	fmt.Println(vr)
}

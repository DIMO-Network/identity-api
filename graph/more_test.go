package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestX(t *testing.T) {
	zeroAddr := common.HexToAddress("")

	xAddr := common.HexToAddress("0xa4")
	yAddr := common.HexToAddress("0xb3")

	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "identity-api").Logger()

	vehicleAddr := common.HexToAddress("0x4e")
	regAddr := common.HexToAddress("0xB9")

	settings := config.Settings{
		DIMORegistryChainID: 1,
		DIMORegistryAddr:    regAddr.Hex(),
		VehicleNFTAddr:      vehicleAddr.Hex(),
	}

	pdb, _ := helpers.StartContainerDatabase(context.TODO(), t, "../migrations")
	contractEventConsumer := services.NewContractsEventsConsumer(pdb, &logger, &settings)
	repo := repositories.New(pdb)

	resolver := NewResolver(repo)
	c := client.New(loader.Middleware(pdb, handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))))

	event := func(name string, source common.Address, args obj) {
		b, _ := json.Marshal(args)

		b, _ = json.Marshal(services.ContractEventData{
			EventName: name,
			Contract:  source,
			Block: services.Block{
				Time: time.Now().Truncate(time.Second),
			},
			Arguments: b,
		})

		err := contractEventConsumer.Process(context.TODO(), &shared.CloudEvent[json.RawMessage]{
			Type:   "zone.dimo.contract.event",
			Source: "chain/1",
			Data:   b,
		})
		if err != nil {
			t.Error("BAD", err)
		}
	}

	event("Transfer", vehicleAddr, obj{"tokenId": 1, "from": zeroAddr, "to": xAddr})
	event("VehicleAttributeSet", regAddr, obj{"tokenId": 1, "attribute": "Model", "info": "Model Y"})
	event("Transfer", vehicleAddr, obj{"tokenId": 1, "from": xAddr, "to": yAddr})
	event("Transfer", vehicleAddr, obj{"tokenId": 2, "from": zeroAddr, "to": yAddr})

	var out any
	c.MustPost(`{
		accessibleVehicles(address: "0x00000000000000000000000000000000000000b3") {
			edges {
				node {
					id
					owner
					model
				}
			}
		}
	}`, &out)

	n := obj{
		"accessibleVehicles": obj{
			"edges": []obj{
				{"node": obj{"id": "2", "owner": yAddr, "model": nil}},
				{"node": obj{"id": "1", "owner": yAddr, "model": "Model Y"}},
			},
		},
	}

	marshalEqual(t, n, out)
}

type obj map[string]any

func marshalEqual(t *testing.T, expected, actual any, msgAndArgs ...any) {
	eb, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("couldn't serialize expected value %v", err)
	}

	ab, err := json.Marshal(actual)
	if err != nil {
		t.Errorf("couldn't serialize actual value %v", err)
	}

	fmt.Println(string(eb))
	fmt.Println(string(ab))

	assert.JSONEq(t, string(eb), string(ab), msgAndArgs...)
}

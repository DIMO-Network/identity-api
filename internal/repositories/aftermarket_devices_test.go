package repositories_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/DIMO-Network/identity-api/graph"
	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/internal/test"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var cloudEvent = shared.CloudEvent[json.RawMessage]{
	ID:          "2SiTVhP3WBhfQQnnnpeBdMR7BSY",
	Source:      "chain/80001",
	SpecVersion: "1.0",
	Subject:     "0x4de1bcf2b7e851e31216fc07989caa902a604784",
	Time:        time.Now(),
	Type:        "zone.dimo.contract.event",
}

var contractEventData = services.ContractEventData{
	ChainID:         80001,
	EventName:       "AftermarketDeviceNodeMinted",
	Contract:        common.HexToAddress("0x4de1bcf2b7e851e31216fc07989caa902a604784"),
	TransactionHash: common.HexToHash("0x811a85e24d0129a2018c9a6668652db63d73bc6d1c76f21b07da2162c6bfea7d"),
	EventSignature:  common.HexToHash("0xd624fd4c3311e1803d230d97ce71fd60c4f658c30a31fbe08edcb211fd90f63f"),
}

var aftermarketDeviceNodeMintedArgs = services.AftermarketDeviceNodeMintedData{
	AftermarketDeviceAddress: common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	ManufacturerID:           big.NewInt(137),
	Owner:                    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	TokenID:                  big.NewInt(42),
}

var aftermarketDevice = models.AftermarketDevice{
	ID:        1,
	Address:   null.BytesFrom(common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes()),
	Owner:     null.BytesFrom(common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes()),
	Serial:    null.StringFrom("aftermarketDeviceSerial-1"),
	Imei:      null.StringFrom("aftermarketDeviceIMEI-1"),
	MintedAt:  null.TimeFrom(time.Now()),
	VehicleID: null.IntFrom(11),
}

var vehicle = models.Vehicle{
	ID:           11,
	OwnerAddress: common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	Make:         null.StringFrom("Ford"),
	Model:        null.StringFrom("Bronco"),
	Year:         null.IntFrom(2022),
	MintedAt:     time.Now(),
}

func createTestServerAndDB(ctx context.Context, t *testing.T, aftermarketDevices []models.AftermarketDevice, vehicles []models.Vehicle) *httptest.Server {
	pdb, _ := test.StartContainerDatabase(ctx, t, test.MigrationsDirRelPath)
	repo := repositories.NewVehiclesRepo(pdb)

	s := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		Repo: repo,
	}}))
	srv := loader.Middleware(pdb, s)

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	app := httptest.NewServer(mux)

	for _, vehicle := range vehicles {
		err := vehicle.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	for _, device := range aftermarketDevices {
		err := device.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	return app

}

func TestOwnedAftermarketDevices(t *testing.T) {
	ctx := context.Background()
	app := createTestServerAndDB(ctx, t, []models.AftermarketDevice{aftermarketDevice}, []models.Vehicle{vehicle})
	defer app.Close()

	r := strings.NewReader(`{
		"query": "query ownedAftermarketDevices($address: Address!){ ownedAftermarketDevices(address: $address) { edges { node { id serial owner address imei mintedAt owner } } }}",
		"operationName": "ownedAftermarketDevices",
		"variables": {
			"address": "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"
		}
}`)

	resp, err := http.Post(app.URL+"/query", "application/json", r)
	if err != nil {
		fmt.Println(err)
	}

	var respBody model.AftermarketDevice
	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(gjson.GetBytes(b, "data.ownedAftermarketDevices.edges.0.node").String()), &respBody)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, strconv.Itoa(aftermarketDevice.ID), respBody.ID)
	assert.Equal(t, common.BytesToAddress(aftermarketDevice.Address.Bytes), *respBody.Address)
	assert.Equal(t, common.BytesToAddress(aftermarketDevice.Owner.Bytes), *respBody.Owner)
	assert.Equal(t, aftermarketDevice.Serial.String, *respBody.Serial)
	assert.Equal(t, aftermarketDevice.Imei.String, *respBody.Imei)
}

func TestOwnedAftermarketDeviceAndLinkedVehicle(t *testing.T) {
	ctx := context.Background()
	app := createTestServerAndDB(ctx, t, []models.AftermarketDevice{aftermarketDevice}, []models.Vehicle{vehicle})
	defer app.Close()

	r := strings.NewReader(`{
		"query": "query ownedAftermarketDevices($address: Address!){ ownedAftermarketDevices(address: $address) { edges { node { id owner address imei serial mintedAt vehicle { id owner make model year mintedAt } } } }}"
			  ,
			  "operationName": "ownedAftermarketDevices",
			  "variables": {
				  "address": "46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"
			  }
	  }`)

	resp, err := http.Post(app.URL+"/query", "application/json", r)
	if err != nil {
		fmt.Println(err)
	}

	var adBody model.AftermarketDevice
	var vehicleBody model.Vehicle
	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(gjson.GetBytes(b, "data.ownedAftermarketDevices.edges.0.node").String()), &adBody)
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(gjson.GetBytes(b, "data.ownedAftermarketDevices.edges.0.node.vehicle").String()), &vehicleBody)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, strconv.Itoa(aftermarketDevice.ID), adBody.ID)
	assert.Equal(t, common.BytesToAddress(aftermarketDevice.Address.Bytes), *adBody.Address)
	assert.Equal(t, common.BytesToAddress(aftermarketDevice.Owner.Bytes), *adBody.Owner)
	assert.Equal(t, aftermarketDevice.Serial.String, *adBody.Serial)
	assert.Equal(t, aftermarketDevice.Imei.String, *adBody.Imei)

	assert.Equal(t, strconv.Itoa(vehicle.ID), vehicleBody.ID)
	assert.Equal(t, common.BytesToAddress(vehicle.OwnerAddress), vehicleBody.Owner)
	assert.Equal(t, common.BytesToAddress(vehicle.OwnerAddress), vehicleBody.Owner)
	assert.Equal(t, vehicle.Make.String, *vehicleBody.Make)
	assert.Equal(t, vehicle.Model.String, *vehicleBody.Model)
	assert.Equal(t, vehicle.Year.Int, *vehicleBody.Year)
	assert.Equal(t, vehicle.MintedAt.UTC().Format(time.RFC1123), vehicleBody.MintedAt.UTC().Format(time.RFC1123))
	assert.Equal(t, vehicle.MintedAt.UTC().Format(time.RFC1123), vehicleBody.MintedAt.UTC().Format(time.RFC1123))
}

func TestAftermarketDeviceNodeMintSingleResponse(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", test.DBSettings.Name).Logger()

	settings := config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}

	config := mocks.NewTestConfig()
	consumer := mocks.NewConsumer(t, config)

	pdb, _ := test.StartContainerDatabase(ctx, t, test.MigrationsDirRelPath)
	contractEventConsumer := services.NewContractsEventsConsumer(pdb, &logger, &settings)

	argBytes, err := json.Marshal(aftermarketDeviceNodeMintedArgs)
	assert.NoError(t, err)

	contractEventData.Arguments = argBytes
	ctEventDataBytes, err := json.Marshal(contractEventData)
	assert.NoError(t, err)

	cloudEvent.Data = ctEventDataBytes
	expectedBytes, err := json.Marshal(cloudEvent)
	assert.NoError(t, err)

	consumer.ExpectConsumePartition(settings.ContractsEventTopic, 0, 0).YieldMessage(&sarama.ConsumerMessage{Value: expectedBytes})

	outputTest, err := consumer.ConsumePartition(settings.ContractsEventTopic, 0, 0)
	assert.NoError(t, err)

	m := <-outputTest.Messages()
	var e shared.CloudEvent[json.RawMessage]
	err = json.Unmarshal(m.Value, &e)
	assert.NoError(t, err)

	err = contractEventConsumer.Process(ctx, &e)
	assert.NoError(t, err)

	ad, err := models.AftermarketDevices(models.AftermarketDeviceWhere.ID.EQ(int(aftermarketDeviceNodeMintedArgs.TokenID.Int64()))).One(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Equal(t, ad.Address.Bytes, aftermarketDeviceNodeMintedArgs.Owner.Bytes())

	adController := repositories.NewVehiclesRepo(pdb)
	res, err := adController.GetOwnedAftermarketDevices(ctx, aftermarketDeviceNodeMintedArgs.Owner, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, *res.Edges[0].Node.Address, aftermarketDeviceNodeMintedArgs.Owner)
}

func TestAftermarketDeviceNodeMintMultiResponse(t *testing.T) {
	ctx := context.Background()

	pdb, _ := test.StartContainerDatabase(ctx, t, test.MigrationsDirRelPath)

	for i := 1; i < 6; i++ {
		ad := models.AftermarketDevice{
			ID:    i,
			Owner: null.BytesFrom(aftermarketDeviceNodeMintedArgs.Owner.Bytes()),
		}

		err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	// 6 5 4 3 2 1
	//     ^
	//     |
	//     after this

	adController := repositories.NewVehiclesRepo(pdb)
	first := 2
	after := "NA==" // 4
	res, err := adController.GetOwnedAftermarketDevices(ctx, aftermarketDeviceNodeMintedArgs.Owner, &first, &after)
	assert.NoError(t, err)

	fmt.Println(res)

	assert.Len(t, res.Edges, 2)
	assert.Equal(t, "3", res.Edges[0].Node.ID)
	assert.Equal(t, "2", res.Edges[1].Node.ID)
}

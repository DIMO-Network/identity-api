package graph

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type VehicleTestSuite struct {
	suite.Suite
	dbCont   testcontainers.Container
	handler  http.Handler
	consumer *services.ContractsEventsConsumer
}

func (s *VehicleTestSuite) SetupSuite() {
	ctx := context.TODO()
	var db db.Store
	db, s.dbCont = helpers.StartContainerDatabase(ctx, s.T(), "../migrations")

	logger := zerolog.Nop()
	vehicleAddr := common.HexToAddress("0x4e")
	regAddr := common.HexToAddress("0xB9")

	settings := config.Settings{
		DIMORegistryChainID: 1,
		DIMORegistryAddr:    regAddr.Hex(),
		VehicleNFTAddr:      vehicleAddr.Hex(),
	}

	repo := base.NewRepository(db, settings)
	resolver := NewResolver(repo)

	s.consumer = services.NewContractsEventsConsumer(db, &logger, &settings)
	s.handler = loader.Middleware(db, handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver})), settings)
}

func (s *VehicleTestSuite) TearDownSuite() {
	s.dbCont.Terminate(context.TODO()) //nolint
}

func (s *VehicleTestSuite) Test_Vehicle(query string, variables map[string]any, expected any) error {
	req := Request{
		Query:     query,
		Variables: variables,
	}

	w := httptest.NewRecorder()
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	r, err := http.NewRequest("POST", "/", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}
	s.handler.ServeHTTP(w, r)

	if w.Code >= http.StatusBadRequest {
		return fmt.Errorf("code %d", w.Code)
	}

	b, err = json.Marshal(expected)
	if err != nil {
		panic(err)
	}

	s.JSONEq(string(b), w.Body.String())

	return nil
}

type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

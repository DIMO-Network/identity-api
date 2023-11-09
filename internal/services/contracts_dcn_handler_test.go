package services

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DCNConsumerTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	settings  config.Settings
	logger    zerolog.Logger
}

func (o *DCNConsumerTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../migrations")

	o.settings = config.Settings{
		DCNRegistryAddr:     contractEventData.Contract.String(),
		DCNResolverAddr:     "0x60627326F55054Ea448e0a7BC750785bD65EF757",
		DIMORegistryChainID: 80001,
	}

	o.logger = zerolog.New(os.Stdout).With().Timestamp().Str("app", test.DBSettings.Name).Logger()
}

// TearDownTest after each test truncate tables
func (s *DCNConsumerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (o *DCNConsumerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestDCNConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(DCNConsumerTestSuite))
}

func (o *DCNConsumerTestSuite) Test_NewNode_Consume_Success() {
	contractEventData.EventName = NewNode.String()
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	var eventData = NewDCNNodeData{
		Node:  test.GenerateDCNNode(),
		Owner: *wallet,
	}

	contractEventConsumer := NewContractsEventsConsumer(o.pdb, &o.logger, &o.settings)
	e := prepareEvent(o.T(), contractEventData, eventData)

	err = contractEventConsumer.Process(o.ctx, &e)
	o.NoError(err)

	dcn, err := models.DCNS().All(o.ctx, o.pdb.DBS().Reader.DB)
	o.NoError(err)

	o.Len(dcn, 1)
	o.Equal(eventData.Node, dcn[0].Node)
	o.Equal(eventData.Owner.Bytes(), dcn[0].OwnerAddress)
}

func (o *DCNConsumerTestSuite) Test_NewDCNExpiration_Consume_Success() {
	contractEventData.EventName = NewExpiration.String()

	_, owner, err := test.GenerateWallet()
	o.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	var eventData = NewDCNExpirationData{
		Node:       test.GenerateDCNNode(),
		Expiration: int(currTime.Unix()),
	}

	d := models.DCN{
		Node:         eventData.Node,
		OwnerAddress: owner.Bytes(),
	}

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	contractEventConsumer := NewContractsEventsConsumer(o.pdb, &o.logger, &o.settings)
	e := prepareEvent(o.T(), contractEventData, eventData)

	err = contractEventConsumer.Process(o.ctx, &e)
	o.NoError(err)

	dcn, err := models.DCNS().All(o.ctx, o.pdb.DBS().Reader.DB)
	o.NoError(err)

	o.Len(dcn, 1)
	o.Equal(eventData.Node, dcn[0].Node)
	o.Equal(owner.Bytes(), dcn[0].OwnerAddress)
	o.Equal(currTime, dcn[0].Expiration.Time)
}

func (o *DCNConsumerTestSuite) Test_DCNNameChanged_Consume_Success() {
	cEventData := contractEventData
	cEventData.EventName = NameChanged.String()
	cEventData.Contract = common.HexToAddress("0x60627326F55054Ea448e0a7BC750785bD65EF757")
	_, owner, err := test.GenerateWallet()
	o.NoError(err)

	newName := "SomeMockName"
	var eventData = DCNNameChangedData{
		Node: test.GenerateDCNNode(),
		Name: newName,
	}

	d := models.DCN{
		Node:         eventData.Node,
		OwnerAddress: owner.Bytes(),
	}

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	contractEventConsumer := NewContractsEventsConsumer(o.pdb, &o.logger, &o.settings)
	e := prepareEvent(o.T(), cEventData, eventData)

	err = contractEventConsumer.Process(o.ctx, &e)
	o.NoError(err)

	dcn, err := models.DCNS().All(o.ctx, o.pdb.DBS().Reader.DB)
	o.NoError(err)

	o.Len(dcn, 1)
	o.Equal(eventData.Node, dcn[0].Node)
	o.Equal(owner.Bytes(), dcn[0].OwnerAddress)
	o.Equal(eventData.Name, dcn[0].Name.String)
}

func (o *DCNConsumerTestSuite) Test_DCN_VehicleIDChanged_Consume_Success() {
	cEventData := contractEventData
	cEventData.EventName = VehicleIdChanged.String()
	cEventData.Contract = common.HexToAddress("0x60627326F55054Ea448e0a7BC750785bD65EF757")

	_, owner, err := test.GenerateWallet()
	o.NoError(err)

	_, owner2, err := test.GenerateWallet()
	o.NoError(err)

	vehicleID := 1
	var eventData = DCNVehicleIdChangedData{
		Node:      test.GenerateDCNNode(),
		VehicleID: big.NewInt(int64(vehicleID)),
	}

	veh := models.Vehicle{
		ID:           vehicleID,
		OwnerAddress: owner2.Bytes(),
	}
	err = veh.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer())
	o.NoError(err)

	d := models.DCN{
		Node:         eventData.Node,
		OwnerAddress: owner.Bytes(),
	}

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	contractEventConsumer := NewContractsEventsConsumer(o.pdb, &o.logger, &o.settings)
	e := prepareEvent(o.T(), cEventData, eventData)

	err = contractEventConsumer.Process(o.ctx, &e)
	o.NoError(err)

	dcn, err := models.DCNS(
		qm.Load(qm.Rels(models.DCNRels.Vehicle)),
	).All(o.ctx, o.pdb.DBS().Reader.DB)
	o.NoError(err)

	o.Len(dcn, 1)
	o.Equal(eventData.Node, dcn[0].Node)
	o.Equal(owner.Bytes(), dcn[0].OwnerAddress)
	o.Equal(vehicleID, dcn[0].R.Vehicle.ID)
}

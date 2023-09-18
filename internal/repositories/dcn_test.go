package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/identity-api/internal/config"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DCNRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container testcontainers.Container
	repo      *Repository
	settings  config.Settings
}

func (o *DCNRepoTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../migrations")

	o.settings = config.Settings{
		DIMORegistryAddr:    "0x2daeF6FF0E2B61aaBF9ADDE1aFDAfDb4f0f1E660",
		DIMORegistryChainID: 80001,
	}
	o.repo = New(o.pdb)
}

// TearDownTest after each test truncate tables
func (s *DCNRepoTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (o *DCNRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", o.container.SessionID())

	if err := o.container.Terminate(o.ctx); err != nil {
		o.T().Fatal(err)
	}
}

// Test Runner
func TestDCNRepoTestSuite(t *testing.T) {
	suite.Run(t, new(DCNRepoTestSuite))
}

func (o *DCNRepoTestSuite) Test_GetDCNByNode_Success() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

	node := test.GenerateDCNNode()
	d := models.DCN{
		Node:         node,
		OwnerAddress: wallet.Bytes(),
		VehicleID:    null.IntFrom(1),
	}

	veh := models.Vehicle{
		ID:           1,
		OwnerAddress: wallet2.Bytes(),
	}
	err = veh.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer())
	o.NoError(err)

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	dcn, err := o.repo.GetDCNByNode(o.ctx, node)
	o.NoError(err)

	o.Equal(dcn.Owner.Bytes(), wallet.Bytes())
	o.Equal(dcn.Node, node)
	o.Equal(*dcn.VehicleID, 1)
}

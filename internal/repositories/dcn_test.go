package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	test "github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
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

type GetDCNs struct {
	Owner     *common.Address
	VehicleID int
	MintedAt  time.Time
	Node      []byte
}

func (o *DCNRepoTestSuite) SetupSuite() {
	o.ctx = context.Background()
	o.pdb, o.container = test.StartContainerDatabase(o.ctx, o.T(), "../../migrations")

	o.settings = config.Settings{
		DIMORegistryAddr:    "0x2daeF6FF0E2B61aaBF9ADDE1aFDAfDb4f0f1E660",
		DIMORegistryChainID: 80001,
	}
	o.repo = New(o.pdb, o.settings)
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
	params := model.DCNBy{
		Node: node,
	}
	veh := models.Vehicle{
		ID:           1,
		OwnerAddress: wallet2.Bytes(),
	}
	err = veh.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer())
	o.NoError(err)

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	dcn, err := o.repo.GetDCNByNode(o.ctx, params.Node)
	o.NoError(err)

	o.Equal(dcn.Owner.Bytes(), wallet.Bytes())
	o.Equal(dcn.Node, node)
	o.Equal(*dcn.VehicleID, 1)
}

func (o *DCNRepoTestSuite) Test_GetDCNByName_Success() {
	_, wallet, err := test.GenerateWallet()
	o.NoError(err)

	_, wallet2, err := test.GenerateWallet()
	o.NoError(err)

	node := test.GenerateDCNNode()
	dcnName := "mockName.dimo"
	d := models.DCN{
		Node:         node,
		OwnerAddress: wallet.Bytes(),
		VehicleID:    null.IntFrom(1),
		Name:         null.StringFrom(dcnName),
	}
	params := model.DCNBy{
		Name: &dcnName,
	}
	veh := models.Vehicle{
		ID:           1,
		OwnerAddress: wallet2.Bytes(),
	}
	err = veh.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer())
	o.NoError(err)

	err = d.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
	o.NoError(err)

	dcn, err := o.repo.GetDCNByName(o.ctx, *params.Name)
	o.NoError(err)

	o.Equal(dcn.Owner.Bytes(), wallet.Bytes())
	o.Equal(dcn.Node, node)
	o.Equal(*dcn.VehicleID, 1)
}

func (o *DCNRepoTestSuite) Test_GetDCNs() {
	// Node    | Owner | Vehicle ID | Minted At
	// --------+-------+-------------------------
	// x1      | A     | 1          | 2023-11-10 20:46:55
	// x2      | A     | 2          | 2023-11-08 20:46:55
	// x3      | B     | 3          | 2023-11-08 20:46:55
	_, walletA, err := test.GenerateWallet()
	o.NoError(err)
	_, walletB, err := test.GenerateWallet()
	o.NoError(err)

	mintedAt, err := time.Parse(time.RFC3339, "2023-11-08T20:46:55Z")
	o.NoError(err)

	data := []GetDCNs{
		{
			Owner:     walletA,
			VehicleID: 1,
			MintedAt:  mintedAt.AddDate(0, 0, 2),
			Node:      []byte{0x1, 170, 166, 15, 216, 74, 75, 44, 94, 252, 64, 71, 50, 35, 193, 212, 53, 140, 163, 41, 192, 80, 127, 168, 193, 96, 213, 8, 56, 94, 182, 22},
		},
		{
			Owner:     walletA,
			VehicleID: 2,
			MintedAt:  mintedAt,
			Node:      []byte{0x2, 170, 166, 15, 216, 74, 75, 44, 94, 252, 64, 71, 50, 35, 193, 212, 53, 140, 163, 41, 192, 80, 127, 168, 193, 96, 213, 8, 56, 94, 182, 22},
		},
		{
			Owner:     walletB,
			VehicleID: 3,
			MintedAt:  mintedAt,
			Node:      []byte{0x3, 170, 166, 15, 216, 74, 75, 44, 94, 252, 64, 71, 50, 35, 193, 212, 53, 140, 163, 41, 192, 80, 127, 168, 193, 96, 213, 8, 56, 94, 182, 22},
		},
	}

	for _, d := range data {
		veh := models.Vehicle{
			ID:           d.VehicleID,
			OwnerAddress: d.Owner.Bytes(),
		}
		err = veh.Insert(o.ctx, o.pdb.DBS().Writer, boil.Infer())
		o.NoError(err)

		dcn := models.DCN{
			Node:         d.Node,
			OwnerAddress: d.Owner.Bytes(),
			VehicleID:    null.IntFrom(d.VehicleID),
			Name:         null.StringFrom(fmt.Sprintf("dcn-%d", d.VehicleID)),
			MintedAt:     d.MintedAt,
		}

		err = dcn.Insert(o.ctx, o.pdb.DBS().Writer.DB, boil.Infer())
		o.NoError(err)
	}

	// first record (ordered by default, DESC)
	// Node    | Owner | Vehicle ID | Minted At
	// --------+-------+-------------------------
	// x1      | A     | 1          | 2023-11-10 20:46:55
	first := 10
	dcnFirst, err := o.repo.GetDCNs(o.ctx, &first, nil, nil, nil, nil)
	o.NoError(err)

	o.Equal(data[0].Owner.Bytes(), dcnFirst.Nodes[0].Owner.Bytes())
	o.Equal(data[0].Node, dcnFirst.Nodes[0].Node)
	o.Equal(data[0].VehicleID, *dcnFirst.Nodes[0].VehicleID)

	// last record (order ASC)
	// Node    | Owner | Vehicle ID | Minted At
	// --------+-------+-------------------------
	// x2      | A     | 2          | 2023-11-08 20:46:55
	last := 1
	dcnLast, err := o.repo.GetDCNs(o.ctx, nil, nil, &last, nil, nil)
	o.NoError(err)

	o.Equal(data[1].Owner.Bytes(), dcnLast.Nodes[0].Owner.Bytes())
	o.Equal(data[1].Node, dcnLast.Nodes[0].Node)
	o.Equal(data[1].VehicleID, *dcnLast.Nodes[0].VehicleID)

	// search after (ordered DESC MintedAt, Node means Token 3 comes before Token 2)
	// Node    | Owner | Vehicle ID | Minted At
	// --------+-------+-------------------------
	// x3      | B     | 3          | 2023-11-08 20:46:55
	pHelp := &helpers.PaginationHelper[DCNCursor]{}
	c, err := pHelp.EncodeCursor(DCNCursor{MintedAt: data[0].MintedAt, Node: data[0].Node})
	o.NoError(err)
	dcnAfter, err := o.repo.GetDCNs(o.ctx, &first, &c, nil, nil, nil)
	o.NoError(err)

	o.Equal(data[2].Owner.Bytes(), dcnAfter.Nodes[0].Owner.Bytes())
	o.Equal(data[2].Node, dcnAfter.Nodes[0].Node)
	o.Equal(data[2].VehicleID, *dcnAfter.Nodes[0].VehicleID)
}

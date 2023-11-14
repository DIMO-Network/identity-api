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

	data := []struct {
		Owner     *common.Address
		VehicleID int
		MintedAt  time.Time
		Node      []byte
	}{
		{
			Owner:     walletA,
			VehicleID: 1,
			MintedAt:  mintedAt.AddDate(0, 0, 2),
			Node:      common.LeftPadBytes(common.FromHex("0x1"), 32),
		},
		{
			Owner:     walletA,
			VehicleID: 2,
			MintedAt:  mintedAt,
			Node:      common.LeftPadBytes(common.FromHex("0x2"), 32),
		},
		{
			Owner:     walletB,
			VehicleID: 3,
			MintedAt:  mintedAt,
			Node:      common.LeftPadBytes(common.FromHex("0x3"), 32),
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

	first := 10
	last := 1
	pHelp := &helpers.PaginationHelper[DCNCursor]{}
	cursor, err := pHelp.EncodeCursor(DCNCursor{MintedAt: mintedAt.AddDate(0, 0, 2), Node: common.LeftPadBytes(common.FromHex("0x1"), 32)})
	o.NoError(err)
	dcnFilter := model.DCNFilter{Owner: walletB}
	for _, testCase := range []struct {
		Description      string
		ExpectedResponse struct {
			Owner     *common.Address
			VehicleID int
			MintedAt  time.Time
			Node      []byte
		}
		First  *int
		Last   *int
		Cursor *string
		Filter *model.DCNFilter
	}{
		{
			// Node    | Owner | Vehicle ID | Minted At
			// --------+-------+-------------------------
			// x1      | A     | 1          | 2023-11-10 20:46:55
			Description:      "first record (ordered by default, DESC)",
			ExpectedResponse: data[0],
			First:            &first,
		},
		{
			// Node    | Owner | Vehicle ID | Minted At
			// --------+-------+-------------------------
			// x2      | A     | 2          | 2023-11-08 20:46:55
			Description:      "last record (order ASC)",
			ExpectedResponse: data[1],
			Last:             &last,
		},
		{
			// Node    | Owner | Vehicle ID | Minted At
			// --------+-------+-------------------------
			// x3      | B     | 3          | 2023-11-08 20:46:55
			Description:      "search after (ordered DESC MintedAt, Node means Token 3 comes before Token 2)",
			ExpectedResponse: data[2],
			First:            &first,
			Cursor:           &cursor,
		},
		{
			// Node    | Owner | Vehicle ID | Minted At
			// --------+-------+-------------------------
			// x3      | B     | 3          | 2023-11-08 20:46:55
			Description:      "filter by owner",
			ExpectedResponse: data[2],
			First:            &first,
			Filter:           &dcnFilter,
		},
	} {
		result, err := o.repo.GetDCNs(o.ctx, testCase.First, testCase.Cursor, testCase.Last, nil, testCase.Filter)
		o.NoError(err)

		o.Equal(testCase.ExpectedResponse.Owner.Bytes(), result.Nodes[0].Owner.Bytes())
		o.Equal(testCase.ExpectedResponse.Node, result.Nodes[0].Node)
		o.Equal(testCase.ExpectedResponse.VehicleID, *result.Nodes[0].VehicleID)
		o.Equal(testCase.ExpectedResponse.MintedAt, result.Nodes[0].MintedAt)
	}
}

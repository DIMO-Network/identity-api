package repositories

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	gmn "github.com/DIMO-Network/go-mnemonic"
	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var aftermarketDeviceNodeMintedArgs = services.AftermarketDeviceNodeMintedData{
	AftermarketDeviceAddress: common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	ManufacturerID:           big.NewInt(137),
	Owner:                    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	TokenID:                  big.NewInt(42),
}

func TestAftermarketDeviceNodeMintMultiResponse(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	for i := 1; i < 6; i++ {
		ad := models.AftermarketDevice{
			ID:          i,
			Owner:       aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Beneficiary: aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Address:     aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
		}

		err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	// 6 5 4 3 2 1
	//     ^
	//     |
	//     after this

	adController := New(pdb)
	first := 2
	after := "NA==" // 4
	res, err := adController.GetAftermarketDevices(ctx, &first, &after, nil, nil, &model.AftermarketDevicesFilter{Owner: &aftermarketDeviceNodeMintedArgs.Owner})
	assert.NoError(t, err)

	fmt.Println(res)

	assert.Len(t, res.Edges, 2)
	assert.Equal(t, 3, res.Edges[0].Node.TokenID)
	assert.Equal(t, 2, res.Edges[1].Node.TokenID)
}

func Test_GetOwnedAftermarketDevices_Pagination_PreviousPage(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	var ad []models.AftermarketDevice
	for i := 1; i <= 4; i++ {
		ad = append(ad, models.AftermarketDevice{
			ID:          i,
			Owner:       aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Beneficiary: aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Address:     aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
		})
	}

	for _, adv := range ad {
		err := adv.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	// 4 3 2 1
	//       ^
	//       |
	//       before this
	adController := New(pdb)
	last := 2
	before := "MQ=="
	startCrsr := "Mw=="
	endCrsr := "Mg=="
	res, err := adController.GetAftermarketDevices(ctx, nil, nil, &last, &before, &model.AftermarketDevicesFilter{Owner: &aftermarketDeviceNodeMintedArgs.Owner})
	assert.NoError(t, err)

	assert.Len(t, res.Edges, 2)
	assert.Equal(t, res.TotalCount, 4)
	assert.Equal(t, res.PageInfo, &model.PageInfo{
		StartCursor:     &startCrsr,
		EndCursor:       &endCrsr,
		HasPreviousPage: true,
		HasNextPage:     true,
	})

	expected := []*model.AftermarketDeviceEdge{
		{
			Node: &model.AftermarketDevice{
				ID:          "AD_kQM=",
				TokenID:     3,
				Owner:       common.BytesToAddress(ad[2].Owner),
				Beneficiary: common.BytesToAddress(ad[2].Beneficiary),
				Address:     common.BytesToAddress(ad[2].Address),
			},
			Cursor: "Mw==",
		},
		{
			Node: &model.AftermarketDevice{
				ID:          "AD_kQI=",
				TokenID:     2,
				Owner:       common.BytesToAddress(ad[1].Owner),
				Beneficiary: common.BytesToAddress(ad[1].Beneficiary),
				Address:     common.BytesToAddress(ad[1].Address),
			},
			Cursor: "Mg==",
		},
	}

	for _, af := range expected {
		mn, err := gmn.EntropyToMnemonicThreeWords(af.Node.Address.Bytes())
		if err != nil {
			assert.NoError(t, err)
		}
		name := strings.Join(mn, " ")

		af.Node.Name = &name
	}
	assert.Exactly(t, expected, res.Edges)
}

package devicedefinition

import (
	"context"
	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"math/big"
	"strings"
	"testing"
)

const migrationsDir = "../../../migrations"

var aftermarketDeviceNodeMintedArgs = services.AftermarketDeviceNodeMintedData{
	AftermarketDeviceAddress: common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	ManufacturerID:           big.NewInt(137),
	Owner:                    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	TokenID:                  big.NewInt(42),
}

func Test_GetDeviceDefinitions_Pagination_PreviousPage(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	mfr := models.Manufacturer{
		ID:      137,
		Name:    "Toyota",
		Owner:   common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:    "toyota",
		TableID: null.IntFrom(1),
	}
	err := mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	var ad []models.AftermarketDevice
	for i := 1; i <= 4; i++ {
		ad = append(ad, models.AftermarketDevice{
			ID:             i,
			ManufacturerID: 137,
			Owner:          aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Beneficiary:    aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Address:        aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
		})
	}

	for _, adv := range ad {
		err := adv.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}
	logger := zerolog.Nop()

	// 4 3 2 1
	//       ^
	//       |
	//       before this
	repo := base.NewRepository(pdb, config.Settings{DIMORegistryChainID: 30001}, &logger)
	adController := Repository{Repository: repo}
	last := 2
	before := "MQ=="
	startCrsr := "Mw=="
	endCrsr := "Mg=="
	res, err := adController.GetDeviceDefinitions(ctx, &mfr.TableID.Int, nil, nil, &last, &before, &model.DeviceDefinitionFilter{Manufacturer: mfr.Slug})
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
				ID:             "AD_kQM=",
				ManufacturerID: 137,
				TokenID:        3,
				Owner:          common.BytesToAddress(ad[2].Owner),
				Beneficiary:    common.BytesToAddress(ad[2].Beneficiary),
				Address:        common.BytesToAddress(ad[2].Address),
				Image:          "https://mockUrl.com/v1/aftermarket/device/3/image",
			},
			Cursor: "Mw==",
		},
		{
			Node: &model.AftermarketDevice{
				ID:             "AD_kQI=",
				TokenID:        2,
				ManufacturerID: 137,
				Owner:          common.BytesToAddress(ad[1].Owner),
				Beneficiary:    common.BytesToAddress(ad[1].Beneficiary),
				Address:        common.BytesToAddress(ad[1].Address),
				Image:          "https://mockUrl.com/v1/aftermarket/device/2/image",
			},
			Cursor: "Mg==",
		},
	}

	for _, af := range expected {
		name := strings.Join(mnemonic.FromInt32WithObfuscation(int32(af.Node.TokenID)), " ")

		af.Node.Name = name
	}
	assert.Exactly(t, expected, res.Edges)
}

//
//func Test_GetDeviceDefinitions_FilterByManufacturer(t *testing.T) {
//	ctx := context.Background()
//	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)
//
//	manufacturerID := 1
//	mnfs := []models.Manufacturer{
//		{
//			ID:       manufacturerID,
//			Owner:    common.HexToAddress("1").Bytes(),
//			MintedAt: time.Now(),
//			Name:     "Toyota",
//			Slug:     "toyota",
//		},
//		{
//			ID:       2,
//			Owner:    common.HexToAddress("2").Bytes(),
//			MintedAt: time.Now(),
//			Name:     "Honda",
//			Slug:     "honda",
//		},
//	}
//	for _, m := range mnfs {
//		err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
//		assert.NoError(t, err)
//	}
//
//	for i := 1; i <= 4; i++ {
//		ad := models.AftermarketDevice{
//			ID:             i,
//			Owner:          aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
//			Beneficiary:    common.BigToAddress(big.NewInt(int64(i + 2))).Bytes(),
//			Address:        aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
//			ManufacturerID: manufacturerID,
//		}
//
//		if i%2 == 0 {
//			ad.ManufacturerID = 2
//		}
//
//		err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer())
//		assert.NoError(t, err)
//	}
//
//	first := 10
//	logger := zerolog.Nop()
//	repo := base.NewRepository(pdb, config.Settings{}, &logger)
//	adController := Repository{Repository: repo}
//	actual, err := adController.GetDeviceDefinitions(ctx, &first, nil, nil, nil, &model.DeviceDefinitionFilter{Manufacturer: &manufacturerID})
//	assert.NoError(t, err)
//
//	assert.Len(t, actual.Edges, 2)
//	assert.Equal(t, actual.TotalCount, 2)
//
//	expected := []struct {
//		id             string
//		manufacturerID int
//		owner          common.Address
//		beneficiary    common.Address
//	}{
//		{
//			id:             "AD_kQM=",
//			manufacturerID: manufacturerID,
//			owner:          aftermarketDeviceNodeMintedArgs.Owner,
//		},
//		{
//			id:             "AD_kQE=",
//			manufacturerID: manufacturerID,
//			owner:          aftermarketDeviceNodeMintedArgs.Owner,
//		},
//	}
//
//	for i, e := range expected {
//		assert.Exactly(t, e.id, actual.Edges[i].Node.ID)
//		assert.Exactly(t, e.manufacturerID, actual.Edges[i].Node.ManufacturerID)
//		assert.Exactly(t, e.owner, actual.Edges[i].Node.Owner)
//	}
//}

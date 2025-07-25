package aftermarket

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/mnemonic"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const migrationsDir = "../../../migrations"

var aftermarketDeviceNodeMintedArgs = services.AftermarketDeviceNodeMintedData{
	AftermarketDeviceAddress: common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	ManufacturerID:           big.NewInt(137),
	Owner:                    common.HexToAddress("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	TokenID:                  big.NewInt(42),
}

func TestAftermarketDeviceNodeMintMultiResponse(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	var mfr2 = models.Manufacturer{
		ID:       137,
		Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		Name:     "AutoPi",
		MintedAt: time.Now(),
		Slug:     "autopi",
	}
	err := mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		ad := models.AftermarketDevice{
			ManufacturerID: 137,
			ID:             i,
			Owner:          aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Beneficiary:    aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Address:        aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
		}

		err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}
	logger := zerolog.Nop()
	// 6 5 4 3 2 1
	//     ^
	//     |
	//     after this
	adController := Repository{Repository: base.NewRepository(pdb, config.Settings{}, &logger)}
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

	mfr := models.Manufacturer{
		ID:    137,
		Name:  "AutoPi",
		Owner: common.FromHex("0xaba3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Slug:  "autopi",
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
	repo := base.NewRepository(pdb, config.Settings{
		BaseImageURL:          "https://mockUrl.com/v1",
		DIMORegistryChainID:   80001,
		AftermarketDeviceAddr: "0x325b45949C833986bC98e98a49F3CA5C5c4643B5",
	}, &logger)
	adController := New(repo)
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
				ID:             "AD_kQM=",
				ManufacturerID: 137,
				TokenID:        3,
				TokenDID:       "did:erc721:80001:0x325b45949C833986bC98e98a49F3CA5C5c4643B5:3",
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
				TokenDID:       "did:erc721:80001:0x325b45949C833986bC98e98a49F3CA5C5c4643B5:2",
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

func Test_GetAftermarketDevices_FilterByBeneficiary(t *testing.T) {
	ctx := context.Background()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	var mfr2 = models.Manufacturer{
		ID:       137,
		Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		Name:     "AutoPi",
		MintedAt: time.Now(),
		Slug:     "autopi",
	}
	err := mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	var tknID int
	// dynamically setting bene for each AD
	// overwriting beneficiary variable each time
	// when we query below, we are expecting only item in response
	var beneficiary common.Address
	for i := 1; i <= 4; i++ {
		tknID = i
		beneficiary = common.BigToAddress(big.NewInt(int64(i + 2)))
		ad := models.AftermarketDevice{
			ID:             tknID,
			ManufacturerID: 137,
			Owner:          aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Beneficiary:    beneficiary.Bytes(),
			Address:        aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
		}

		err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	expectedBeneResp := &model.AftermarketDeviceEdge{
		Node: &model.AftermarketDevice{
			TokenID:     tknID,
			Owner:       aftermarketDeviceNodeMintedArgs.Owner,
			Beneficiary: beneficiary,
			Address:     aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress,
		},
	}

	first := 10
	logger := zerolog.Nop()
	adController := New(base.NewRepository(pdb, config.Settings{}, &logger))
	beneFilterRes, err := adController.GetAftermarketDevices(ctx, &first, nil, nil, nil, &model.AftermarketDevicesFilter{Beneficiary: &beneficiary})
	assert.NoError(t, err)

	assert.Len(t, beneFilterRes.Edges, 1)
	assert.Equal(t, beneFilterRes.TotalCount, 1)
	assert.Exactly(t, expectedBeneResp.Node.TokenID, beneFilterRes.Edges[0].Node.TokenID)
	assert.Exactly(t, expectedBeneResp.Node.Owner, beneFilterRes.Edges[0].Node.Owner)
	assert.Exactly(t, expectedBeneResp.Node.Beneficiary, beneFilterRes.Edges[0].Node.Beneficiary)
	assert.Exactly(t, expectedBeneResp.Node.Address, beneFilterRes.Edges[0].Node.Address)

	// changing filter to owner will return all ADs
	ownerFilterResp, err := adController.GetAftermarketDevices(ctx, &first, nil, nil, nil, &model.AftermarketDevicesFilter{Owner: &aftermarketDeviceNodeMintedArgs.Owner})
	assert.NoError(t, err)
	assert.Len(t, ownerFilterResp.Edges, 4)
	assert.Equal(t, ownerFilterResp.TotalCount, 4)
}

func Test_GetAftermarketDevices_FilterByManufacturerID(t *testing.T) {
	ctx := context.Background()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	manufacturerID := 1
	mnfs := []models.Manufacturer{
		{
			ID:       manufacturerID,
			Owner:    common.HexToAddress("1").Bytes(),
			MintedAt: time.Now(),
			Name:     "Toyota",
			Slug:     "toyota",
		},
		{
			ID:       2,
			Owner:    common.HexToAddress("2").Bytes(),
			MintedAt: time.Now(),
			Name:     "Honda",
			Slug:     "honda",
		},
	}
	for _, m := range mnfs {
		err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	for i := 1; i <= 4; i++ {
		ad := models.AftermarketDevice{
			ID:             i,
			Owner:          aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
			Beneficiary:    common.BigToAddress(big.NewInt(int64(i + 2))).Bytes(),
			Address:        aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
			ManufacturerID: manufacturerID,
		}

		if i%2 == 0 {
			ad.ManufacturerID = 2
		}

		err := ad.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	first := 10
	logger := zerolog.Nop()
	repo := base.NewRepository(pdb, config.Settings{}, &logger)
	adController := Repository{Repository: repo}
	actual, err := adController.GetAftermarketDevices(ctx, &first, nil, nil, nil, &model.AftermarketDevicesFilter{ManufacturerID: &manufacturerID})
	assert.NoError(t, err)

	assert.Len(t, actual.Edges, 2)
	assert.Equal(t, actual.TotalCount, 2)

	expected := []struct {
		id             string
		manufacturerID int
		owner          common.Address
		beneficiary    common.Address
	}{
		{
			id:             "AD_kQM=",
			manufacturerID: manufacturerID,
			owner:          aftermarketDeviceNodeMintedArgs.Owner,
		},
		{
			id:             "AD_kQE=",
			manufacturerID: manufacturerID,
			owner:          aftermarketDeviceNodeMintedArgs.Owner,
		},
	}

	for i, e := range expected {
		assert.Exactly(t, e.id, actual.Edges[i].Node.ID)
		assert.Exactly(t, e.manufacturerID, actual.Edges[i].Node.ManufacturerID)
		assert.Exactly(t, e.owner, actual.Edges[i].Node.Owner)
	}
}

func Test_GetAftermarketDeviceImageUrl(t *testing.T) {
	testCases := []struct {
		name        string
		baseURL     string
		tokenID     int
		expectedURL string
	}{
		{
			name:        "valid url",
			baseURL:     "https://mockUrl.com/v1",
			tokenID:     42,
			expectedURL: "https://mockUrl.com/v1/aftermarket/device/42/image",
		},
		{
			name:        "empty url",
			baseURL:     "",
			tokenID:     42,
			expectedURL: "aftermarket/device/42/image",
		},
		{
			name:        "leading slash",
			baseURL:     "/v1",
			tokenID:     42,
			expectedURL: "/v1/aftermarket/device/42/image",
		},
		{
			name:        "escaped base url",
			baseURL:     "<div>",
			tokenID:     42,
			expectedURL: "%3Cdiv%3E/aftermarket/device/42/image",
		},
		{
			name:        "trailing slash",
			baseURL:     "https://mockUrl.com/v1/",
			tokenID:     42,
			expectedURL: "https://mockUrl.com/v1/aftermarket/device/42/image",
		},
	}

	for _, tc := range testCases {
		url, err := GetAftermarketDeviceImageURL(tc.baseURL, tc.tokenID)
		require.NoError(t, err)
		require.Equal(t, tc.expectedURL, url)
	}
}

func TestAftermarketDeviceBy(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	var mfr2 = models.Manufacturer{
		ID:       137,
		Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		Name:     "AutoPi",
		MintedAt: time.Now(),
		Slug:     "autopi",
	}
	err := mfr2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	imeiAd := models.AftermarketDevice{
		ManufacturerID: 137,
		ID:             1,
		Owner:          aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
		Beneficiary:    aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
		Address:        aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
		Imei:           null.StringFrom("123456789012345"),
	}

	err = imeiAd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	devEUIAd := models.AftermarketDevice{
		ManufacturerID: 137,
		ID:             2,
		Owner:          aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
		Beneficiary:    aftermarketDeviceNodeMintedArgs.Owner.Bytes(),
		Address:        aftermarketDeviceNodeMintedArgs.AftermarketDeviceAddress.Bytes(),
		DevEui:         null.StringFrom("123456789012345"),
	}

	err = devEUIAd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	logger := zerolog.Nop()
	adController := New(base.NewRepository(pdb, config.Settings{}, &logger))
	res, err := adController.GetAftermarketDevice(ctx, model.AftermarketDeviceBy{Imei: ref("123456789012345")})
	require.NoError(t, err)
	require.Equal(t, imeiAd.ID, res.TokenID)

	res, err = adController.GetAftermarketDevice(ctx, model.AftermarketDeviceBy{DevEui: ref("123456789012345")})
	require.NoError(t, err)
	require.Equal(t, devEUIAd.ID, res.TokenID)
}

func ref[T any](v T) *T { return &v }

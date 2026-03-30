package graph

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/loader"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type testVehicleDefinitionFetcher struct {
	docs map[string]*services.DeviceDefinitionDoc
}

func (t testVehicleDefinitionFetcher) GetVehicleDefinitionDoc(_ context.Context, vehicleDID string) (*services.DeviceDefinitionDoc, error) {
	return t.docs[vehicleDID], nil
}

func TestVehicleDefinitionFetchConsistencyAcrossQueryShapes(t *testing.T) {
	ctx := context.Background()
	pdb, cont := helpers.StartContainerDatabase(ctx, t, migrationsDir)
	defer cont.Terminate(t.Context()) //nolint

	logger := zerolog.Nop()
	settings := config.Settings{
		DIMORegistryChainID: 1,
		DIMORegistryAddr:    common.HexToAddress("0xB9").Hex(),
		VehicleNFTAddr:      common.HexToAddress("0x4e").Hex(),
	}

	mfr := models.Manufacturer{
		ID:       41,
		Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Name:     "Ford",
		MintedAt: time.Now(),
		Slug:     "ford",
	}
	require.NoError(t, mfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	adMfr := models.Manufacturer{
		ID:       137,
		Owner:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDff"),
		Name:     "AutoPi",
		MintedAt: time.Now(),
		Slug:     "autopi",
	}
	require.NoError(t, adMfr.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	v := models.Vehicle{
		ID:             11,
		ManufacturerID: 41,
		OwnerAddress:   common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		Make:           null.StringFrom("Ford"),
		Model:          null.StringFrom("Bronco"),
		Year:           null.IntFrom(2022),
		MintedAt:       time.Now(),
	}
	require.NoError(t, v.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	ad := models.AftermarketDevice{
		ID:             1,
		ManufacturerID: 137,
		Address:        common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf5").Bytes(),
		Owner:          common.HexToAddress("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4").Bytes(),
		Serial:         null.StringFrom("aftermarketDeviceSerial-1"),
		Imei:           null.StringFrom("aftermarketDeviceIMEI-1"),
		MintedAt:       time.Now(),
		VehicleID:      null.IntFrom(11),
		Beneficiary:    common.FromHex("46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
	}
	require.NoError(t, ad.Insert(ctx, pdb.DBS().Writer, boil.Infer()))

	repo := base.NewRepository(pdb, settings, &logger)
	resolver := NewResolver(repo)

	vehicleDID := cloudevent.ERC721DID{
		ChainID:         uint64(settings.DIMORegistryChainID),
		ContractAddress: common.HexToAddress(settings.VehicleNFTAddr),
		TokenID:         big.NewInt(11),
	}.String()
	fetcher := testVehicleDefinitionFetcher{
		docs: map[string]*services.DeviceDefinitionDoc{
			vehicleDID: {
				ID:    "fetch-definition-id",
				Make:  "Fetch Ford",
				Model: "Fetch Bronco",
				Year:  2025,
			},
		},
	}
	resolver.vehicleDefFetch = fetcher

	c := client.New(loader.MiddlewareWithFetcher(
		pdb,
		NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver})),
		settings,
		&logger,
		fetcher,
	))

	var topLevelResp struct {
		Vehicle struct {
			Definition struct {
				ID    *string `json:"id"`
				Make  *string `json:"make"`
				Model *string `json:"model"`
				Year  *int    `json:"year"`
			} `json:"definition"`
		} `json:"vehicle"`
	}
	c.MustPost(`{ vehicle(tokenId: 11) { definition { id make model year } } }`, &topLevelResp)
	require.Equal(t, "fetch-definition-id", *topLevelResp.Vehicle.Definition.ID)
	require.Equal(t, "Fetch Ford", *topLevelResp.Vehicle.Definition.Make)
	require.Equal(t, "Fetch Bronco", *topLevelResp.Vehicle.Definition.Model)
	require.Equal(t, 2025, *topLevelResp.Vehicle.Definition.Year)

	var nestedResp struct {
		AftermarketDevice struct {
			Vehicle struct {
				Definition struct {
					ID    *string `json:"id"`
					Make  *string `json:"make"`
					Model *string `json:"model"`
					Year  *int    `json:"year"`
				} `json:"definition"`
			} `json:"vehicle"`
		} `json:"aftermarketDevice"`
	}
	query := fmt.Sprintf(`{ aftermarketDevice(by: {tokenId: %d}) { vehicle { definition { id make model year } } } }`, ad.ID)
	c.MustPost(query, &nestedResp)
	require.Equal(t, "fetch-definition-id", *nestedResp.AftermarketDevice.Vehicle.Definition.ID)
	require.Equal(t, "Fetch Ford", *nestedResp.AftermarketDevice.Vehicle.Definition.Make)
	require.Equal(t, "Fetch Bronco", *nestedResp.AftermarketDevice.Vehicle.Definition.Model)
	require.Equal(t, 2025, *nestedResp.AftermarketDevice.Vehicle.Definition.Year)
}

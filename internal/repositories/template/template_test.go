package template

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const migrationsDir = "../../../migrations"

func TestGetTemplate(t *testing.T) {
	ctx := context.Background()

	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	var tokenIds [3]*big.Int
	creator := common.HexToAddress("3232323232323232323232323232323232323232")
	asset := common.HexToAddress("5454545454545454545454545454545454545454")
	permissions := new(big.Int).SetUint64(3888).Text(2) // 11 11 00 11 00 00
	cids := []string{"ford", "tesla", "kia"}

	for i := range 3 {
		tokenId := helpers.StringToUint256Hash(cids[i])
		tokenIds[i] = tokenId

		tokenIdBytes, err := helpers.ConvertTokenIDToID(tokenId)
		assert.NoError(t, err)

		m := models.Template{
			ID:          tokenIdBytes,
			Creator:     creator.Bytes(),
			Asset:       asset.Bytes(),
			Permissions: permissions,
			CreatedAt:   time.Now(),
			Cid:         cids[i],
		}

		err = m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	logger := zerolog.Nop()
	controller := New(base.NewRepository(pdb, config.Settings{}, &logger))
	for i := range 3 {
		res, err := controller.GetTemplate(ctx, model.TemplateBy{TokenID: tokenIds[i]})
		assert.NoError(t, err)
		assert.Equal(t, res.TokenID, tokenIds[i])
		t.Logf("res.Creator type: %T, value: %v", res.Creator, res.Creator)
		t.Logf("creator type: %T, value: %v", creator, creator)
		assert.Equal(t, res.Creator, creator)
		assert.Equal(t, res.Asset, asset)
		assert.Equal(t, res.Permissions, permissions)
		assert.Equal(t, res.Cid, cids[i])

		res, err = controller.GetTemplate(ctx, model.TemplateBy{Cid: &cids[i]})
		assert.NoError(t, err)
		assert.Equal(t, res.TokenID, tokenIds[i])
		assert.Equal(t, res.Creator, creator)
		assert.Equal(t, res.Asset, asset)
		assert.Equal(t, res.Permissions, permissions)
		assert.Equal(t, res.Cid, cids[i])
	}
}

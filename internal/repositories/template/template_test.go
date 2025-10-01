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
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const migrationsDir = "../../../migrations"

func TestGetTemplate(t *testing.T) {
	ctx := context.Background()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	tokenIds, creators, assets, permissions, cids := createTemplates(t, ctx, pdb)

	logger := zerolog.Nop()
	controller := New(base.NewRepository(pdb, config.Settings{}, &logger))
	for i := range 3 {
		res, err := controller.GetTemplate(ctx, model.TemplateBy{TokenID: tokenIds[i]})
		assert.NoError(t, err)
		assert.Equal(t, res.TokenID, tokenIds[i])
		assert.Equal(t, res.Creator, creators[i])
		assert.Equal(t, res.Asset, assets[i])
		assert.Equal(t, res.Permissions, permissions)
		assert.Equal(t, res.Cid, cids[i])

		res, err = controller.GetTemplate(ctx, model.TemplateBy{Cid: &cids[i]})
		assert.NoError(t, err)
		assert.Equal(t, res.TokenID, tokenIds[i])
		assert.Equal(t, res.Creator, creators[i])
		assert.Equal(t, res.Asset, assets[i])
		assert.Equal(t, res.Permissions, permissions)
		assert.Equal(t, res.Cid, cids[i])
	}
}

func TestGetTemplates(t *testing.T) {
	ctx := context.Background()
	pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)

	tokenIds, creators, assets, permissions, cids := createTemplates(t, ctx, pdb)

	logger := zerolog.Nop()
	controller := New(base.NewRepository(pdb, config.Settings{}, &logger))

	first := 100
	res, err := controller.GetTemplates(ctx, &first, nil, nil, nil)
	assert.NoError(t, err)

	totalTemplates := len(tokenIds)

	assert.Len(t, res.Nodes, totalTemplates)
	assert.Equal(t, res.TotalCount, totalTemplates)
	assert.Equal(t, res.PageInfo.HasNextPage, false)
	assert.Equal(t, res.PageInfo.HasPreviousPage, false)
	assert.Len(t, res.Edges, totalTemplates)
	for i := range totalTemplates {
		// GetTemplates returns descending order
		reverseIndex := totalTemplates - 1 - i

		assert.Equal(t, res.Nodes[i].TokenID, tokenIds[reverseIndex])
		assert.Equal(t, res.Nodes[i].Creator, creators[reverseIndex])
		assert.Equal(t, res.Nodes[i].Asset, assets[reverseIndex])
		assert.Equal(t, res.Nodes[i].Permissions, permissions)
		assert.Equal(t, res.Nodes[i].Cid, cids[reverseIndex])

		assert.Equal(t, res.Edges[i].Node.TokenID, tokenIds[reverseIndex])
		assert.Equal(t, res.Edges[i].Node.Creator, creators[reverseIndex])
		assert.Equal(t, res.Edges[i].Node.Asset, assets[reverseIndex])
		assert.Equal(t, res.Edges[i].Node.Permissions, permissions)
		assert.Equal(t, res.Edges[i].Node.Cid, cids[reverseIndex])
	}
}

func createTemplates(t *testing.T, ctx context.Context, pdb db.Store) ([]*big.Int, []common.Address, []common.Address, string, []string) {
	size := 3

	tokenIds := make([]*big.Int, size)
	creators := []common.Address{
		common.HexToAddress("1111111111111111111111111111111111111111"),
		common.HexToAddress("2222222222222222222222222222222222222222"),
		common.HexToAddress("3333333333333333333333333333333333333333"),
	}
	assets := []common.Address{
		common.HexToAddress("5555555555555555555555555555555555555555"),
		common.HexToAddress("6666666666666666666666666666666666666666"),
		common.HexToAddress("7777777777777777777777777777777777777777"),
	}
	permissions := new(big.Int).SetUint64(3888).Text(2) // 11 11 00 11 00 00
	cids := []string{
		"QmTkDGhbhABxgHde1Go2vcarW3NxQYkaSHFBYCdiUo79d9",
		"QmdYe42GrGU9TCtN4rq1K1bsa5naaiZUhBV93PrF4JtYaE",
		"QmYoWDRp6yXc53rZBaWrz56XEaDm7heoDNx6s5ttESiXHX",
	}

	for i := range size {
		tokenId := helpers.StringToUint256Hash(cids[i])
		tokenIds[i] = tokenId

		tokenIdBytes, err := helpers.ConvertTokenIDToID(tokenId)
		assert.NoError(t, err)

		m := models.Template{
			ID:          tokenIdBytes,
			Creator:     creators[i].Bytes(),
			Asset:       assets[i].Bytes(),
			Permissions: permissions,
			CreatedAt:   time.Now(),
			Cid:         cids[i],
		}

		err = m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err)
	}

	return tokenIds, creators, assets, permissions, cids
}

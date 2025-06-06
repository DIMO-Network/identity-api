package loader

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/graph-gophers/dataloader/v7"
)

type ManufacturerLoader struct {
	repo *manufacturer.Repository
}

func NewManufacturerLoader(repo *manufacturer.Repository) *ManufacturerLoader {
	return &ManufacturerLoader{repo: repo}
}

func GetManufacturerID(ctx context.Context, manufacturerID int) (*model.Manufacturer, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.ManufacturerByID.Load(ctx, manufacturerID)
	// read value from thunk
	return thunk()
}

func (m *ManufacturerLoader) BatchGetManufacturerByID(ctx context.Context, manufacturerIDs []int) []*dataloader.Result[*model.Manufacturer] {
	results := make([]*dataloader.Result[*model.Manufacturer], len(manufacturerIDs))

	manufacturers, err := models.Manufacturers(models.ManufacturerWhere.ID.IN(manufacturerIDs)).All(ctx, m.repo.PDB.DBS().Reader)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*model.Manufacturer]{Error: err}
		}
		return results
	}

	manufacturerByID := map[int]*models.Manufacturer{}

	for _, v := range manufacturers {
		manufacturerByID[v.ID] = v
	}

	for i, k := range manufacturerIDs {
		if v, ok := manufacturerByID[k]; ok {
			obj, err := m.repo.ToAPI(v)
			results[i] = &dataloader.Result[*model.Manufacturer]{
				Data:  obj,
				Error: err,
			}
		} else {
			results[i] = &dataloader.Result[*model.Manufacturer]{Error: fmt.Errorf("no manufacturer with id %d", k)}
		}
	}

	return results
}

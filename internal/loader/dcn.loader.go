package loader

import (
	"context"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/dcn"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/graph-gophers/dataloader/v7"
)

type DCNLoader struct {
	db db.Store
}

func GetDCNByVehicleID(ctx context.Context, vehicleID int) (*gmodel.Dcn, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.DCNByVehicleID.Load(ctx, vehicleID)
	// read value from thunk
	return thunk()
}

// BatchGetDCNByVehicleID implements the dataloader for finding DCN linked to vehicles and returns
// them in the order requested
func (d *DCNLoader) BatchGetDCNByVehicleID(ctx context.Context, vehicleIDs []int) []*dataloader.Result[*gmodel.Dcn] {
	results := make([]*dataloader.Result[*gmodel.Dcn], len(vehicleIDs))
	dcns, err := models.DCNS(models.DCNWhere.VehicleID.IN(vehicleIDs)).All(ctx, d.db.DBS().Reader)
	if err != nil {
		for i := range vehicleIDs {
			results[i] = &dataloader.Result[*gmodel.Dcn]{Data: nil, Error: err}
		}
		return results
	}

	dcnByVehicleID := map[int]*models.DCN{}
	for _, d := range dcns {
		dcnByVehicleID[d.VehicleID.Int] = d
	}

	for idx, vid := range vehicleIDs {
		if dcnModel, ok := dcnByVehicleID[vid]; ok {
			dcnAPI, err := dcn.ToAPI(dcnModel)
			results[idx] = &dataloader.Result[*gmodel.Dcn]{
				Data:  dcnAPI,
				Error: err,
			}
		} else {
			results[idx] = &dataloader.Result[*gmodel.Dcn]{}
		}

	}

	return results
}

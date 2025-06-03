package loader

import (
	"context"
	"errors"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/connection"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ConnectionLoader struct {
	db db.Store
}

func GetConnectionByID(ctx context.Context, id [32]byte) (*model.Connection, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)
	// invoke and get thunk
	thunk := loaders.ConnectionByID.Load(ctx, id)
	// read value from thunk
	return thunk()
}

func (ad *ConnectionLoader) BatchGetConnectionsByIDs(ctx context.Context, ids [][32]byte) []*dataloader.Result[*model.Connection] {
	boil.DebugMode = true

	results := make([]*dataloader.Result[*model.Connection], len(ids))

	uniqIDs := make(map[[32]byte]struct{})

	for _, id := range ids {
		uniqIDs[id] = struct{}{}
	}

	queryIDs := make([][]byte, 0, len(uniqIDs))
	for id := range uniqIDs {
		queryIDs = append(queryIDs, id[:])
	}

	connections, err := models.Connections(qm.Where(models.ConnectionTableColumns.ID+" = ANY(?)", pq.ByteaArray(queryIDs))).All(ctx, ad.db.DBS().Reader)
	if err != nil {
		for i := range ids {
			results[i] = &dataloader.Result[*model.Connection]{Data: nil, Error: err}
		}
		return results
	}

	connectionByID := map[[32]byte]*models.Connection{}

	for _, d := range connections {
		connectionByID[[32]byte(d.ID)] = d
	}

	for i, vID := range ids {
		// We're okay with the missing case here. We just want to return something for now.
		am, ok := connectionByID[vID]
		if ok {
			results[i] = &dataloader.Result[*model.Connection]{
				Data: connection.ToAPI(am),
			}
		} else {
			results[i] = &dataloader.Result[*model.Connection]{
				Error: errors.New("couldn't find a connection with that id"),
			}
		}
	}

	return results
}

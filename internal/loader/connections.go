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
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ConnectionLoader struct {
	db db.Store
}

type ConnectionQueryKey struct {
	IntegrationNode int
	ConnectionID    [32]byte
}

func GetConnectionByID(ctx context.Context, integrationNode int, connectionID []byte) (*model.Connection, error) {
	// read loader from context
	loaders := ctx.Value(dataLoadersKey).(*Loaders)

	query := ConnectionQueryKey{
		IntegrationNode: integrationNode,
	}

	if len(connectionID) == 32 {
		query.ConnectionID = [32]byte(connectionID)
	}

	// invoke and get thunk
	thunk := loaders.ConnectionByID.Load(ctx, query)
	// read value from thunk
	return thunk()
}

func (ad *ConnectionLoader) BatchGetConnectionsByIDs(ctx context.Context, ids []ConnectionQueryKey) []*dataloader.Result[*model.Connection] {
	results := make([]*dataloader.Result[*model.Connection], len(ids))

	uniqKeys := make(map[ConnectionQueryKey]struct{})

	for _, id := range ids {
		uniqKeys[id] = struct{}{}
	}

	connectionIDs := make([][]byte, 0)
	integrationNodes := make([]int32, 0)

	for id := range uniqKeys {
		if id.IntegrationNode == 0 {
			connectionIDs = append(connectionIDs, id.ConnectionID[:])
		} else {
			integrationNodes = append(integrationNodes, int32(id.IntegrationNode))
		}
	}

	connections, err := models.Connections(
		qm.Where(models.ConnectionTableColumns.ID+" = ANY(?)", pq.ByteaArray(connectionIDs)),
		qm.Or(models.ConnectionTableColumns.IntegrationNode+" = ANY(?)", pq.Int32Array(integrationNodes)),
	).All(ctx, ad.db.DBS().Reader)
	if err != nil {
		for i := range ids {
			results[i] = &dataloader.Result[*model.Connection]{Data: nil, Error: err}
		}
		return results
	}

	connectionByConnectionID := map[[32]byte]*models.Connection{}
	connectionByIntegrationNode := map[int]*models.Connection{}

	for _, d := range connections {
		connectionByConnectionID[[32]byte(d.ID)] = d
		if d.IntegrationNode.Valid {
			connectionByIntegrationNode[d.IntegrationNode.Int] = d
		}
	}

	for i, vID := range ids {
		if vID.IntegrationNode == 0 {
			am, ok := connectionByConnectionID[vID.ConnectionID]
			if ok {
				results[i] = &dataloader.Result[*model.Connection]{
					Data: connection.ToAPI(am),
				}
			} else {
				results[i] = &dataloader.Result[*model.Connection]{
					Error: errors.New("couldn't find a connection with that id"),
				}
			}
		} else {
			am, ok := connectionByIntegrationNode[vID.IntegrationNode]
			if ok {
				results[i] = &dataloader.Result[*model.Connection]{
					Data: connection.ToAPI(am),
				}
			} else {
				results[i] = &dataloader.Result[*model.Connection]{
					Error: errors.New("couldn't find a connection corresponding to that integration node"),
				}
			}
		}
	}

	return results
}

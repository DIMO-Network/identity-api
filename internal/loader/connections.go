package loader

import (
	"context"
	"errors"
	"fmt"

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

type ConnectionQueryKey struct {
	ConnectionID    [32]byte
	IntegrationNode int
}

// GetConnection retrieves a connection based on either an integration node or connection id.
// Either the integrationNode is nonzero or the connectionID is length 32, but not both.
func GetConnection(ctx context.Context, integrationNode int, connectionID []byte) (*model.Connection, error) {
	loaders := ctx.Value(dataLoadersKey).(*Loaders)

	queryKey := ConnectionQueryKey{
		IntegrationNode: integrationNode,
	}

	if len(connectionID) == 32 {
		queryKey.ConnectionID = [32]byte(connectionID)
	}

	// invoke and get thunk
	thunk := loaders.ConnectionByID.Load(ctx, queryKey)
	// read value from thunk
	return thunk()
}

func (ad *ConnectionLoader) BatchGetConnectionsByIDs(ctx context.Context, queryKeys []ConnectionQueryKey) []*dataloader.Result[*model.Connection] {
	uniqKeys := make(map[ConnectionQueryKey]struct{})

	for _, id := range queryKeys {
		uniqKeys[id] = struct{}{}
	}

	fmt.Println("XDD", queryKeys)

	uniqueConnectionIDs := make(map[[32]byte]struct{})
	uniqueIntegrationNodes := make(map[int]struct{})

	for key := range uniqKeys {
		if key.IntegrationNode != 0 {
			uniqueIntegrationNodes[key.IntegrationNode] = struct{}{}
		} else {
			uniqueConnectionIDs[key.ConnectionID] = struct{}{}
		}
	}

	boil.DebugMode = true
	results := make([]*dataloader.Result[*model.Connection], len(queryKeys))

	connectionIDSlice := make([][]byte, 0, len(uniqueConnectionIDs))
	for connID := range uniqueConnectionIDs {
		connectionIDSlice = append(connectionIDSlice, connID[:])
	}

	integrationNodeSlice := make([]int32, 0, len(uniqueIntegrationNodes))
	for intNode := range uniqueIntegrationNodes {
		integrationNodeSlice = append(integrationNodeSlice, int32(intNode))
	}

	connections, err := models.Connections(
		qm.Where(models.ConnectionTableColumns.ID+" = ANY(?)", pq.ByteaArray(connectionIDSlice)),
		qm.Or(models.ConnectionTableColumns.IntegrationNode+" = ANY(?)", pq.Int32Array(integrationNodeSlice)),
	).All(ctx, ad.db.DBS().Reader)
	if err != nil {
		for i := range queryKeys {
			results[i] = &dataloader.Result[*model.Connection]{Data: nil, Error: err}
		}
		return results
	}

	fmt.Println(connections)

	connectionByConnectionID := map[[32]byte]*models.Connection{}
	connectionByIntegrationNode := map[int]*models.Connection{}

	for _, d := range connections {
		connectionByConnectionID[[32]byte(d.ID)] = d
		if d.IntegrationNode.Valid {
			connectionByIntegrationNode[d.IntegrationNode.Int] = d
		}
	}

	for i, queryKey := range queryKeys {
		if queryKey.IntegrationNode != 0 {
			am, ok := connectionByIntegrationNode[queryKey.IntegrationNode]
			if ok {
				results[i] = &dataloader.Result[*model.Connection]{
					Data: connection.ToAPI(am),
				}
			} else {
				results[i] = &dataloader.Result[*model.Connection]{
					Error: errors.New("couldn't find a connection corresponding to that integration node"),
				}
			}
		} else {
			am, ok := connectionByConnectionID[queryKey.ConnectionID]
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
	}

	return results
}

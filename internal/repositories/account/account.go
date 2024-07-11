package account

import (
	"context"
	"fmt"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

// AccountPrefix is the prefix for a global token id for Accounts.
const AccountPrefix = "Acct"

type Repository struct {
	*base.Repository
}

type AccountCursor struct {
	OwnerAddress []byte
	Kernel       []byte
}

// ToAPI converts a Account database row to a Account API model.
func ToAPI(d *models.KernelAccount) (*gmodel.Account, error) {
	return &gmodel.Account{
		Kernel: []common.Address{common.BytesToAddress(d.Kernel)},
		Owner:  common.BytesToAddress(d.OwnerAddress),
	}, nil
}

var accountCursorColumnsTuple = "(" + models.KernelAccountColumns.OwnerAddress + ", " + models.KernelAccountColumns.Kernel + ")"

func (r *Repository) GetAccount(ctx context.Context, by gmodel.AccountBy) (*gmodel.Account, error) {
	if by.Owner == nil {
		return nil, gqlerror.Errorf("Provide exactly one of `owner`.")
	}

	return r.GetAccountByOwner(ctx, *by.Owner)
}

func (r *Repository) GetAccountByOwner(ctx context.Context, owner common.Address) (*gmodel.Account, error) {
	acc, err := models.KernelAccounts(models.KernelAccountWhere.OwnerAddress.EQ(owner.Bytes())).One(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}
	return ToAPI(acc)
}

func (r *Repository) GetAccounts(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.AccountFilter) (*gmodel.AccountConnection, error) {
	fmt.Println("Getting lots of accounts: ", first, after, last, before, filterBy)
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{}
	if filterBy != nil && filterBy.Owner != nil {
		queryMods = append(queryMods, models.KernelAccountWhere.OwnerAddress.EQ(filterBy.Owner.Bytes()))
	}

	accountCount, err := models.KernelAccounts(queryMods...).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	orderBy := " DESC"
	if last != nil {
		orderBy = " ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.KernelAccountColumns.OwnerAddress+orderBy+", "+models.KernelAccountColumns.Kernel+orderBy),
	)

	pHelp := &helpers.PaginationHelper[AccountCursor]{}
	if after != nil {
		afterT, err := pHelp.DecodeCursor(*after)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods,
			qm.Where(accountCursorColumnsTuple+" < (?, ?)", afterT.OwnerAddress, afterT.Kernel),
		)
	} else if before != nil {
		beforeT, err := pHelp.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods,
			qm.Where(accountCursorColumnsTuple+" < (?, ?)", beforeT.OwnerAddress, beforeT.Kernel),
		)
	}

	all, err := models.KernelAccounts(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	hasNext := before != nil
	hasPrevious := after != nil

	if first != nil && len(all) == limit+1 {
		hasNext = true
		all = all[:limit]
	} else if last != nil && len(all) == limit+1 {
		hasPrevious = true
		all = all[:limit]
	}

	if last != nil {
		slices.Reverse(all)
	}

	edges := make([]*gmodel.AccountEdge, len(all))
	nodes := make([]*gmodel.Account, len(all))
	var errList gqlerror.List
	for i, acct := range all {
		c, err := pHelp.EncodeCursor(AccountCursor{OwnerAddress: acct.OwnerAddress, Kernel: acct.Kernel})
		if err != nil {
			return nil, err
		}
		apiAccount, err := ToAPI(acct)
		if err != nil {
			errList = append(errList, gqlerror.Errorf("error converting account to api: %v", err))
			continue
		}
		edges[i] = &gmodel.AccountEdge{
			Node:   apiAccount,
			Cursor: c,
		}
		nodes[i] = edges[i].Node
	}

	var endCur, startCur *string

	if len(all) != 0 {
		startCur = &edges[0].Cursor
		endCur = &edges[len(edges)-1].Cursor
	}

	res := &gmodel.AccountConnection{
		TotalCount: int(accountCount),
		Edges:      edges,
		Nodes:      nodes,
		PageInfo: &gmodel.PageInfo{
			StartCursor:     startCur,
			EndCursor:       endCur,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
		},
	}
	if errList != nil {
		return res, errList
	}
	return res, nil
}

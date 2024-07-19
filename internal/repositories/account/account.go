package account

import (
	"context"

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
	Signer []byte
	Kernel []byte
}

var accountCursorColumnsTuple = "(" + models.KernelAccountColumns.OwnerAddress + ", " + models.KernelAccountColumns.Kernel + ")"

func (r *Repository) GetAccount(ctx context.Context, by gmodel.AccountBy) (*gmodel.Account, error) {

	var acctBy []qm.QueryMod
	switch {
	case by.Signer != nil || by.Address != nil:
		acctBy = append(acctBy, models.KernelAccountWhere.OwnerAddress.EQ(by.Signer.Bytes()))
	case by.Kernel != nil || by.Address != nil:
		acctBy = append(acctBy, models.KernelAccountWhere.Kernel.EQ(by.Kernel.Bytes()))
	default:
		return nil, gqlerror.Errorf("Provide exactly one of `signer`, `kernel` or `address`.")
	}

	all, err := models.KernelAccounts(acctBy...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	var apiAccount *gmodel.Account
	for _, acct := range all {
		if apiAccount == nil {
			apiAccount = &gmodel.Account{
				Signer: common.BytesToAddress(acct.OwnerAddress),
			}
		}

		apiAccount.Kernel = append(apiAccount.Kernel, &gmodel.Kernel{
			Address:   common.BytesToAddress(acct.Kernel),
			CreatedAt: acct.CreatedAt,
			Signer: &gmodel.Signer{
				Address:     common.BytesToAddress(acct.OwnerAddress),
				SignerAdded: acct.SignerAdded,
			},
		})
	}

	return apiAccount, nil
}

func (r *Repository) GetAccounts(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *gmodel.AccountFilter) (*gmodel.AccountConnection, error) {
	limit, err := helpers.ValidateFirstLast(first, last, base.MaxPageSize)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{}
	if filterBy != nil && filterBy.Signer != nil {
		queryMods = append(queryMods, models.KernelAccountWhere.OwnerAddress.EQ(filterBy.Signer.Bytes()))
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
			qm.Where(accountCursorColumnsTuple+" < (?, ?)", afterT.Signer, afterT.Kernel),
		)
	} else if before != nil {
		beforeT, err := pHelp.DecodeCursor(*before)
		if err != nil {
			return nil, err
		}
		queryMods = append(queryMods,
			qm.Where(accountCursorColumnsTuple+" < (?, ?)", beforeT.Signer, beforeT.Kernel),
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

	ordered := 0
	ownerToKernals := make(map[common.Address]AcctInfos)
	for _, acct := range all {
		owner := common.BytesToAddress(acct.OwnerAddress)
		c, err := pHelp.EncodeCursor(AccountCursor{Signer: acct.OwnerAddress, Kernel: acct.Kernel})
		if err != nil {
			return nil, err
		}

		acctInfos, ok := ownerToKernals[owner]
		if !ok {
			acctInfos = AcctInfos{Idx: ordered}
			ordered++
		}

		acctInfos.Kernels = append(acctInfos.Kernels, &gmodel.Kernel{
			Address:   common.BytesToAddress(acct.Kernel),
			CreatedAt: acct.CreatedAt,
			Signer: &gmodel.Signer{
				Address:     common.BytesToAddress(owner.Bytes()),
				SignerAdded: acct.SignerAdded,
			},
		})
		acctInfos.Cursor = append(acctInfos.Cursor, c)
		ownerToKernals[owner] = acctInfos

	}

	edges := make([]*gmodel.AccountEdge, len(ownerToKernals))
	nodes := make([]*gmodel.Account, len(ownerToKernals))
	var errList gqlerror.List

	for owner, acctInfos := range ownerToKernals {
		apiAccount := &gmodel.Account{
			Signer: owner,
			Kernel: acctInfos.Kernels,
		}

		edges[acctInfos.Idx] = &gmodel.AccountEdge{
			Node: apiAccount,
		}

		if len(acctInfos.Cursor) != 0 {
			edges[acctInfos.Idx].Cursor = acctInfos.Cursor[len(acctInfos.Cursor)-1]
		}
		nodes[acctInfos.Idx] = edges[acctInfos.Idx].Node
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

type AcctInfos struct {
	Idx     int
	Kernels []*gmodel.Kernel
	Cursor  []string
}

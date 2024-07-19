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
	Signer []byte
	Kernel []byte
}

var accountCursorColumnsTuple = "(" + models.AccountColumns.Signer + ", " + models.AccountColumns.Kernel + ")"

func (r *Repository) GetAccount(ctx context.Context, by gmodel.AccountBy) (*gmodel.Account, error) {
	var acctBy []qm.QueryMod

	switch {
	case by.Signer != nil && by.Kernel == nil && by.Address == nil:
		acctBy = append(acctBy, models.AccountWhere.Signer.EQ(by.Signer.Bytes()))
	case by.Kernel != nil && by.Address == nil && by.Signer == nil:
		acctBy = append(acctBy, models.AccountWhere.Kernel.EQ(by.Kernel.Bytes()))
	case by.Address != nil && by.Signer == nil && by.Kernel == nil:
		acctBy = append(acctBy, qm.Or2(models.AccountWhere.Kernel.EQ(by.Address.Bytes())), qm.Or2(models.AccountWhere.Signer.EQ(by.Address.Bytes())))
	default:
		return nil, gqlerror.Errorf("Provide exactly one of `signer`, `kernel` or `address`.")
	}

	all, err := models.Accounts(acctBy...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		r.Log.Err(err).Msg("failed fetching account")
		return nil, fmt.Errorf("failed fetching account")
	}

	var apiAccount *gmodel.Account
	for _, acct := range all {
		if apiAccount == nil {
			apiAccount = &gmodel.Account{
				Signer: common.BytesToAddress(acct.Signer),
			}
		}

		apiAccount.Kernel = append(apiAccount.Kernel, &gmodel.Kernel{
			Address:   common.BytesToAddress(acct.Kernel),
			CreatedAt: acct.CreatedAt,
			Signer: &gmodel.Signer{
				Address:     common.BytesToAddress(acct.Signer),
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
		queryMods = append(queryMods, models.AccountWhere.Signer.EQ(filterBy.Signer.Bytes()))
	}

	accountCount, err := models.Accounts(queryMods...).Count(ctx, r.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	orderBy := " DESC"
	if last != nil {
		orderBy = " ASC"
	}

	queryMods = append(queryMods,
		qm.Limit(limit+1),
		qm.OrderBy(models.AccountColumns.Signer+orderBy),
		qm.OrderBy(models.AccountColumns.Kernel+orderBy),
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

	all, err := models.Accounts(queryMods...).All(ctx, r.PDB.DBS().Reader)
	if err != nil {
		r.Log.Err(err).Msg("failed fetching account")
		return nil, fmt.Errorf("failed fetching account")
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
	signerToKernels := make(map[common.Address]AcctInfos)
	for _, acct := range all {
		signer := common.BytesToAddress(acct.Signer)
		c, err := pHelp.EncodeCursor(AccountCursor{Signer: acct.Signer, Kernel: acct.Kernel})
		if err != nil {
			return nil, err
		}

		acctInfos, ok := signerToKernels[signer]
		if !ok {
			acctInfos = AcctInfos{Idx: ordered}
			ordered++
		}

		acctInfos.Kernels = append(acctInfos.Kernels, &gmodel.Kernel{
			Address:   common.BytesToAddress(acct.Kernel),
			CreatedAt: acct.CreatedAt,
			Signer: &gmodel.Signer{
				Address:     common.BytesToAddress(signer.Bytes()),
				SignerAdded: acct.SignerAdded,
			},
		})
		acctInfos.Cursor = append(acctInfos.Cursor, c)
		signerToKernels[signer] = acctInfos

	}

	edges := make([]*gmodel.AccountEdge, len(signerToKernels))
	nodes := make([]*gmodel.Account, len(signerToKernels))
	for signer, acctInfos := range signerToKernels {
		apiAccount := &gmodel.Account{
			Signer: signer,
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

	return res, nil
}

type AcctInfos struct {
	Idx     int
	Kernels []*gmodel.Kernel
	Cursor  []string
}

package vehiclesacd

import (
	"context"
	"errors"

	gmodel "github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/ethereum/go-ethereum/common"
)

type Repository struct {
	*base.Repository
}

func ToAPI(dl *models.DeveloperLicense) (*gmodel.DeveloperLicense, error) {
	return &gmodel.DeveloperLicense{
		TokenID:  dl.TokenID,
		Owner:    common.BytesToAddress(dl.Owner),
		ClientID: common.BytesToAddress(dl.ClientID),
		MintedAt: dl.MintedAt,
		Alias:    dl.Alias.Ptr(),
	}, nil
}

func (p *Repository) GetDeveloperLicense(ctx context.Context, by gmodel.DeveloperLicenseBy) (*gmodel.DeveloperLicense, error) {
	if by.ClientID == nil {
		return nil, errors.New("must provide a client id address")
	}

	dl, err := models.DeveloperLicenses(models.DeveloperLicenseWhere.ClientID.EQ(by.ClientID.Bytes())).One(ctx, p.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return ToAPI(dl)
}

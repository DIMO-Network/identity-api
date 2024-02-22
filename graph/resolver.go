package graph

import (
	"github.com/DIMO-Network/identity-api/internal/repositories"
	"github.com/DIMO-Network/identity-api/internal/repositories/aftermarket"
	"github.com/DIMO-Network/identity-api/internal/repositories/dcn"
	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"
	"github.com/DIMO-Network/identity-api/internal/repositories/reward"
	"github.com/DIMO-Network/identity-api/internal/repositories/synthetic"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicleprivilege"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	aftermarket      aftermarket.Repository
	dcn              dcn.Repository
	manufacturer     manufacturer.Repository
	reward           reward.Repository
	synthetic        synthetic.Repository
	vehicle          vehicle.Repository
	vehicleprivilege vehicleprivilege.Repository
}

func NewResolver(repo *repositories.Repository) *Resolver {
	return &Resolver{
		aftermarket:      aftermarket.Repository{Repository: repo},
		dcn:              dcn.Repository{Repository: repo},
		manufacturer:     manufacturer.Repository{Repository: repo},
		reward:           reward.Repository{Repository: repo},
		synthetic:        synthetic.Repository{Repository: repo},
		vehicle:          vehicle.Repository{Repository: repo},
		vehicleprivilege: vehicleprivilege.Repository{Repository: repo},
	}

}

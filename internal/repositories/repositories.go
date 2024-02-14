package repositories

import (
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/shared/db"
)

const (
	// MaxPageSize is the maximum page size for paginated results
	MaxPageSize = 100
)

type Repository struct {
	PDB      db.Store
	Settings config.Settings
}

func New(pdb db.Store, settings config.Settings) *Repository {
	return &Repository{
		PDB:      pdb,
		Settings: settings,
	}
}

func CountTrue(ps ...bool) int {
	out := 0

	for _, p := range ps {
		if p {
			out++
		}
	}

	return out
}

package services

import (
	"context"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/patrickmn/go-cache"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/rs/zerolog"
)

type ManufacturerCacheModel struct {
	ID   int
	Name string
	Slug string
}

type ManufacturerCacheService struct {
	log      *zerolog.Logger
	settings *config.Settings
	cache    *cache.Cache
	PDB      db.Store
}

func NewManufacturerCacheService(pdb db.Store,
	log *zerolog.Logger,
	settings *config.Settings) *ManufacturerCacheService {

	return &ManufacturerCacheService{
		PDB:      pdb,
		log:      log,
		settings: settings,
		cache:    cache.New(0, 0),
	}
}

func (m *ManufacturerCacheService) GetAllManufacturers(ctx context.Context) ([]ManufacturerCacheModel, error) {
	if manufacturers, ok := m.cache.Get("manufacturers"); ok {
		return manufacturers.([]ManufacturerCacheModel), nil
	}

	manufacturers, err := models.Manufacturers().All(ctx, m.PDB.DBS().Reader)
	if err != nil {
		return nil, err
	}

	var all []ManufacturerCacheModel
	all = make([]ManufacturerCacheModel, 0)
	for _, manufacturer := range manufacturers {
		all = append(all, ManufacturerCacheModel{
			ID:   manufacturer.ID,
			Name: manufacturer.Name,
			Slug: helpers.SlugString(manufacturer.Name),
		})
	}

	m.cache.Set("manufacturers", all, time.Hour*24*7)

	return all, nil
}

package loader

import (
	"context"
	"sync"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/rs/zerolog"
)

const vehicleDefinitionFetchConcurrency = 8

type VehicleDefinitionFetcher interface {
	GetVehicleDefinitionDoc(ctx context.Context, vehicleDID string) (*services.DeviceDefinitionDoc, error)
}

func overlayVehicleDefinition(def *model.Definition, doc *services.DeviceDefinitionDoc) *model.Definition {
	if def == nil {
		def = &model.Definition{}
	}
	if doc == nil {
		return def
	}
	if doc.ID != "" {
		def.ID = &doc.ID
	}
	if doc.Make != "" {
		def.Make = &doc.Make
	}
	if doc.Model != "" {
		def.Model = &doc.Model
	}
	if doc.Year != 0 {
		def.Year = &doc.Year
	}

	return def
}

func enrichVehicleDefinitions(ctx context.Context, log *zerolog.Logger, fetcher VehicleDefinitionFetcher, vehicles []*model.Vehicle) {
	if fetcher == nil || len(vehicles) == 0 {
		return
	}

	sem := make(chan struct{}, vehicleDefinitionFetchConcurrency)
	var wg sync.WaitGroup

	for _, vehicle := range vehicles {
		if vehicle == nil || vehicle.TokenDID == "" {
			continue
		}

		wg.Add(1)
		go func(vehicle *model.Vehicle) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			doc, err := fetcher.GetVehicleDefinitionDoc(ctx, vehicle.TokenDID)
			if err != nil {
				if log != nil {
					log.Warn().Err(err).Str("vehicleDID", vehicle.TokenDID).Msg("fetch-api definition lookup failed, using DB values")
				}
				return
			}

			vehicle.Definition = overlayVehicleDefinition(vehicle.Definition, doc)
		}(vehicle)
	}

	wg.Wait()
}

func EnrichVehicleDefinitions(ctx context.Context, log *zerolog.Logger, fetcher VehicleDefinitionFetcher, vehicles []*model.Vehicle) {
	enrichVehicleDefinitions(ctx, log, fetcher, vehicles)
}

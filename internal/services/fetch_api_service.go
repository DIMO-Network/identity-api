package services

import (
	"context"
	"encoding/json"
	"fmt"

	fetchpb "github.com/DIMO-Network/identity-api/internal/fetchapi"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// DeviceDefinitionDoc holds the device definition fields stored in the fetch-api.
type DeviceDefinitionDoc struct {
	ID    string
	Make  string
	Model string
	Year  int
}

type ddManufacturer struct {
	Name string `json:"name"`
}

type ddPayload struct {
	Data struct {
		DeviceDefinitionID string         `json:"deviceDefinitionId"`
		Manufacturer       ddManufacturer `json:"manufacturer"`
		Model              string         `json:"model"`
		Year               int            `json:"year"`
	} `json:"data"`
}

// FetchAPIService wraps the fetch-api gRPC client.
type FetchAPIService struct {
	client fetchpb.FetchServiceClient
	log    *zerolog.Logger
}

// NewFetchAPIService creates a new FetchAPIService connecting to the given gRPC address.
func NewFetchAPIService(addr string, log *zerolog.Logger) (*FetchAPIService, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to fetch-api at %s: %w", addr, err)
	}
	return &FetchAPIService{
		client: fetchpb.NewFetchServiceClient(conn),
		log:    log,
	}, nil
}

// GetVehicleDefinitionDoc queries the fetch-api for the latest dimo.document.devicedefinition
// cloud event for the given vehicle DID. Returns nil, nil when no document exists.
func (s *FetchAPIService) GetVehicleDefinitionDoc(ctx context.Context, vehicleDID string) (*DeviceDefinitionDoc, error) {
	resp, err := s.client.GetLatestCloudEvent(ctx, &fetchpb.GetLatestCloudEventRequest{
		Options: &fetchpb.SearchOptions{
			Subject: wrapperspb.String(vehicleDID),
			Type:    wrapperspb.String("dimo.document.devicedefinition"),
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("fetch-api GetLatestCloudEvent for %s: %w", vehicleDID, err)
	}

	if resp.GetCloudEvent().GetData() == nil {
		return nil, nil
	}

	var payload ddPayload
	if err := json.Unmarshal(resp.GetCloudEvent().GetData(), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal device definition doc for %s: %w", vehicleDID, err)
	}

	return &DeviceDefinitionDoc{
		ID:    payload.Data.DeviceDefinitionID,
		Make:  payload.Data.Manufacturer.Name,
		Model: payload.Data.Model,
		Year:  payload.Data.Year,
	}, nil
}

package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type ManufacturerContractService struct {
	log              *zerolog.Logger
	settings         *config.Settings
	registryInstance *contracts.Registry
}

func NewManufacturerContractService(log *zerolog.Logger,
	settings *config.Settings,
	client *ethclient.Client) (*ManufacturerContractService, error) {

	contractAddress := common.HexToAddress(settings.DIMORegistryAddr)
	registryInstance, err := contracts.NewRegistry(contractAddress, client)

	if err != nil {
		return nil, fmt.Errorf("failed instance NewRegistry: %s", err)
	}

	return &ManufacturerContractService{
		log:              log,
		settings:         settings,
		registryInstance: registryInstance,
	}, nil
}

func (m *ManufacturerContractService) GetTableName(ctx context.Context, manufacturerID int) (*string, error) {

	tableName, err := m.registryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{
		Context: ctx,
		Pending: true,
	}, big.NewInt(int64(manufacturerID)))

	if err != nil {
		return nil, gqlerror.Errorf("failed get GetDeviceDefinitionTableName: %s", err)
	}

	return &tableName, nil
}

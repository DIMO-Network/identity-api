package services

import (
	"context"

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
	log      *zerolog.Logger
	settings *config.Settings
	client   *ethclient.Client
}

func NewManufacturerContractService(log *zerolog.Logger,
	settings *config.Settings,
	client *ethclient.Client) *ManufacturerContractService {
	return &ManufacturerContractService{
		log:      log,
		settings: settings,
		client:   client,
	}
}

func (m *ManufacturerContractService) GetTableName(ctx context.Context, manufacturerID int) (*string, error) {
	contractAddress := common.HexToAddress(m.settings.DIMORegistryAddr)
	queryInstance, err := contracts.NewRegistry(contractAddress, m.client)
	if err != nil {
		return nil, gqlerror.Errorf("failed instance NewRegistry: %s", err)
	}

	tableName, err := queryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{
		Context: ctx,
		Pending: true,
	}, big.NewInt(int64(manufacturerID)))

	if err != nil {
		return nil, gqlerror.Errorf("failed get GetDeviceDefinitionTableName: %s", err)
	}

	return &tableName, nil
}

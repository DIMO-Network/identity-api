package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ContractsEventsConsumer struct {
	dbs      db.Store
	log      *zerolog.Logger
	settings *config.Settings
}

type EventName string

var zeroAddress common.Address

const (
	Transfer                           EventName = "Transfer"
	VehicleAttributeSet                EventName = "VehicleAttributeSet"
	AftermarketDeviceAttributeSetEvent EventName = "AftermarketDeviceAttributeSet"
	PrivilegeSet                       EventName = "PrivilegeSet"
	AftermarketDevicePairedEvent       EventName = "AftermarketDevicePaired"
	AftermarketDeviceUnpairedEvent     EventName = "AftermarketDeviceUnpaired"
	BeneficiarySetEvent                EventName = "BeneficiarySet"
	AftermarketDeviceNodeMintedEvent   EventName = "AftermarketDeviceNodeMinted"
	SyntheticDeviceNodeMinted          EventName = "SyntheticDeviceNodeMinted"
	SyntheticDeviceNodeBurned          EventName = "SyntheticDeviceNodeBurned"
)

func (r EventName) String() string {
	return string(r)
}

const contractEventCEType = "zone.dimo.contract.event"

func NewContractsEventsConsumer(dbs db.Store, log *zerolog.Logger, settings *config.Settings) *ContractsEventsConsumer {
	return &ContractsEventsConsumer{
		dbs:      dbs,
		log:      log,
		settings: settings,
	}
}

func (c *ContractsEventsConsumer) Process(ctx context.Context, event *shared.CloudEvent[json.RawMessage]) error {
	// Filter out end-of-block events.
	if event.Type != contractEventCEType {
		return nil
	}

	if event.Source != fmt.Sprintf("chain/%d", c.settings.DIMORegistryChainID) {
		return nil
	}

	registryAddr := common.HexToAddress(c.settings.DIMORegistryAddr)
	vehicleNFTAddr := common.HexToAddress(c.settings.VehicleNFTAddr)
	aftermarketDeviceAddr := common.HexToAddress(c.settings.AftermarketDeviceAddr)

	var data ContractEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	eventName := EventName(data.EventName)

	c.log.Info().Str("Event", string(eventName)).Str("Contract", data.Contract.Hex()).Msg("Event Received")

	switch data.Contract {
	case registryAddr:
		switch eventName {
		case VehicleAttributeSet:
			return c.handleVehicleAttributeSetEvent(ctx, &data)
		case AftermarketDeviceNodeMintedEvent:
			return c.handleAftermarketDeviceMintedEvent(ctx, &data)
		case AftermarketDeviceAttributeSetEvent:
			return c.handleAftermarketDeviceAttributeSetEvent(ctx, &data)
		case AftermarketDevicePairedEvent:
			return c.handleAftermarketDevicePairedEvent(ctx, &data)
		case AftermarketDeviceUnpairedEvent:
			return c.handleAftermarketDeviceUnpairedEvent(ctx, &data)
		case BeneficiarySetEvent:
			return c.handleBeneficiarySetEvent(ctx, &data)
		case SyntheticDeviceNodeMinted:
			return c.handleSyntheticDeviceNodeMintedEvent(ctx, &data)
		case SyntheticDeviceNodeBurned:
			return c.handleSyntheticDeviceNodeBurnedEvent(ctx, &data)
		}
	case vehicleNFTAddr:
		switch eventName {
		case Transfer:
			return c.handleVehicleTransferEvent(ctx, &data)
		case PrivilegeSet:
			return c.handlePrivilegeSetEvent(ctx, &data)
		}
	case aftermarketDeviceAddr:
		switch eventName {
		case Transfer:
			return c.handleAftermarketDeviceTransferredEvent(ctx, &data)
		}
	}

	c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceMintedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDeviceNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:      int(args.TokenID.Int64()),
		Address: null.BytesFrom(args.AftermarketDeviceAddress.Bytes()),
	}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.Address))
	return err
}

func (c *ContractsEventsConsumer) handleVehicleAttributeSetEvent(ctx context.Context, e *ContractEventData) error {
	var args VehicleAttributeSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	veh, err := models.FindVehicle(ctx, c.dbs.DBS().Reader, int(args.TokenID.Int64()))
	if err != nil {
		return err
	}

	switch args.Attribute {
	case "Make", "Model", "Year":
		return fmt.Errorf("ignoring MMY attributes %q", args.Attribute)
	case "Definition URI":
		res, err := http.Get(args.Info)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("error occurred fetching device Definition for device with token %d: %d", args.TokenID.Int64(), res.StatusCode)
		}

		ddf := DeviceDefinition{}
		err = json.NewDecoder(res.Body).Decode(&ddf)
		if err != nil {
			return errors.Wrap(err, "failed to marshal json from device definition")
		}

		veh.Make = null.StringFrom(ddf.Type.Make)
		veh.Model = null.StringFrom(ddf.Type.Model)
		veh.Year = null.IntFrom(ddf.Type.Year)
		veh.DefinitionURI = null.StringFrom(args.Info)

		_, err = veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.Make, models.VehicleColumns.Model, models.VehicleColumns.Year, models.VehicleColumns.DefinitionURI))

		return err
	default:
		return fmt.Errorf("unrecognized vehicle attribute %q", args.Attribute)
	}

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleTransferEvent(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", Transfer.String()).Logger()

	var args TransferEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	vehicle := models.Vehicle{
		ID:           int(args.TokenID.Int64()),
		OwnerAddress: args.To.Bytes(),
		MintedAt:     e.Block.Time,
	}

	// Insert is the mint case.
	if err := vehicle.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		true,
		[]string{models.VehicleColumns.ID},
		boil.Whitelist(models.VehicleColumns.OwnerAddress),
		boil.Whitelist(models.VehicleColumns.ID, models.VehicleColumns.OwnerAddress, models.VehicleColumns.MintedAt)); err != nil {
		return err
	}

	if _, err := models.Privileges(models.PrivilegeWhere.TokenID.EQ(int(args.TokenID.Int64()))).DeleteAll(ctx, c.dbs.DBS().Writer); err != nil {
		return err
	}

	logger.Info().Str("TokenID", args.TokenID.String()).Msg("Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceAttributeSetEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDeviceAttributeSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID: int(args.TokenID.Int64()),
	}

	switch args.Attribute {
	case "Serial":
		ad.Serial = null.StringFrom(args.Info)
		if _, err := ad.Update(
			ctx,
			c.dbs.DBS().Writer,
			boil.Whitelist(models.AftermarketDeviceColumns.Serial)); err != nil {
			return err
		}
	case "IMEI":
		ad.Imei = null.StringFrom(args.Info)
		if _, err := ad.Update(
			ctx,
			c.dbs.DBS().Writer,
			boil.Whitelist(models.AftermarketDeviceColumns.Imei)); err != nil {
			return err
		}
	}

	return nil
}

func (c *ContractsEventsConsumer) handlePrivilegeSetEvent(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", PrivilegeSet.String()).Logger()

	var args PrivilegeSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	privilege := models.Privilege{
		TokenID:     int(args.TokenId.Int64()),
		PrivilegeID: int(args.PrivId.Int64()),
		UserAddress: args.User.Bytes(),
		SetAt:       e.Block.Time,
		ExpiresAt:   time.Unix(args.Expires.Int64(), 0),
	}

	if err := privilege.Upsert(ctx, c.dbs.DBS().Writer, true,
		[]string{
			models.PrivilegeColumns.PrivilegeID,
			models.PrivilegeColumns.TokenID,
			models.PrivilegeColumns.UserAddress,
		},
		boil.Whitelist(models.PrivilegeColumns.SetAt, models.PrivilegeColumns.ExpiresAt),
		boil.Infer()); err != nil {
		return err
	}

	logger.Info().
		Str("PrivilegeID", args.PrivId.String()).
		Str("TokenID", args.TokenId.String()).
		Msg("Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDevicePairedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:        int(args.AftermarketDeviceNode.Int64()),
		VehicleID: null.IntFrom(int(args.VehicleNode.Int64())),
	}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleID))
	return err
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceUnpairedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{ID: int(args.AftermarketDeviceNode.Int64())}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleID))
	return err
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceTransferredEvent(ctx context.Context, e *ContractEventData) error {
	var args TransferEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:          int(args.TokenID.Int64()),
		Owner:       args.To.Bytes(),
		MintedAt:    e.Block.Time,
		Beneficiary: args.To.Bytes(),
	}

	return ad.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		true,
		[]string{models.AftermarketDeviceColumns.ID},
		boil.Whitelist(models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.Beneficiary),
		boil.Whitelist(models.AftermarketDeviceColumns.ID, models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.MintedAt, models.AftermarketDeviceColumns.Beneficiary),
	)
}

func (c *ContractsEventsConsumer) handleBeneficiarySetEvent(ctx context.Context, e *ContractEventData) error {
	var args BeneficiarySetEventData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	if args.IdProxyAddress != common.HexToAddress(c.settings.AftermarketDeviceAddr) {
		c.log.Warn().Msgf("beneficiary set on an unexpected contract: %s", args.IdProxyAddress.Hex())
		return nil
	}

	ad := &models.AftermarketDevice{ID: int(args.NodeId.Int64())}

	if args.Beneficiary == zeroAddress {
		if err := ad.Reload(ctx, c.dbs.DBS().Reader); err != nil {
			return err
		}
		ad.Beneficiary = ad.Owner
	} else {
		ad.Beneficiary = args.Beneficiary.Bytes()
	}

	if _, err := ad.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.AftermarketDeviceColumns.Beneficiary),
	); err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleSyntheticDeviceNodeMintedEvent(ctx context.Context, e *ContractEventData) error {
	var args SyntheticDeviceNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	sd := models.SyntheticDevice{
		ID:            int(args.SyntheticDeviceNode.Int64()),
		IntegrationID: int(args.IntegrationNode.Int64()),
		VehicleID:     int(args.VehicleNode.Int64()),
		DeviceAddress: args.SyntheticDeviceAddress.Bytes(),
		MintedAt:      e.Block.Time,
	}

	return sd.Insert(ctx, c.dbs.DBS().Writer, boil.Infer())
}

func (c *ContractsEventsConsumer) handleSyntheticDeviceNodeBurnedEvent(ctx context.Context, e *ContractEventData) error {
	var args SyntheticDeviceNodeBurnedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	sd := models.SyntheticDevice{
		ID: int(args.SyntheticDeviceNode.Int64()),
	}

	_, err := sd.Delete(ctx, c.dbs.DBS().Writer)
	return err
}

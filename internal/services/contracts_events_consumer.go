package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ContractsEventsConsumer struct {
	dbs        db.Store
	log        *zerolog.Logger
	settings   *config.Settings
	httpClient *http.Client
}

type EventName string

var zeroAddress common.Address

const (
	Transfer                      EventName = "Transfer"
	VehicleAttributeSet           EventName = "VehicleAttributeSet"
	ManufacturerNodeMinted        EventName = "ManufacturerNodeMinted"
	AftermarketDeviceAttributeSet EventName = "AftermarketDeviceAttributeSet"
	PrivilegeSet                  EventName = "PrivilegeSet"
	AftermarketDeviceClaimed      EventName = "AftermarketDeviceClaimed"
	AftermarketDevicePaired       EventName = "AftermarketDevicePaired"
	AftermarketDeviceUnpaired     EventName = "AftermarketDeviceUnpaired"
	BeneficiarySetEvent           EventName = "BeneficiarySet"
	VehicleNodeMinted             EventName = "VehicleNodeMinted"
	AftermarketDeviceNodeMinted   EventName = "AftermarketDeviceNodeMinted"
	SyntheticDeviceNodeMinted     EventName = "SyntheticDeviceNodeMinted"
	SyntheticDeviceNodeBurned     EventName = "SyntheticDeviceNodeBurned"
	NewNode                       EventName = "NewNode"
	NewExpiration                 EventName = "NewExpiration"
	NameChanged                   EventName = "NameChanged"
	VehicleIdChanged              EventName = "VehicleIdChanged"
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
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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
	DCNRegistryAddr := common.HexToAddress(c.settings.DCNRegistryAddr)
	DCNResolverAddr := common.HexToAddress(c.settings.DCNResolverAddr)

	var data ContractEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	eventName := EventName(data.EventName)

	c.log.Info().Str("Event", string(eventName)).Str("Contract", data.Contract.Hex()).Msg("Event Received")

	switch data.Contract {
	case registryAddr:
		switch eventName {
		case ManufacturerNodeMinted:
			return c.handleManufacturerNodeMintedEvent(ctx, &data)

		case VehicleNodeMinted:
			return c.handleVehicleNodeMintedEvent(ctx, &data)
		case VehicleAttributeSet:
			return c.handleVehicleAttributeSetEvent(ctx, &data)

		case AftermarketDeviceNodeMinted:
			return c.handleAftermarketDeviceMintedEvent(ctx, &data)
		case AftermarketDeviceAttributeSet:
			return c.handleAftermarketDeviceAttributeSetEvent(ctx, &data)
		case AftermarketDeviceClaimed:
			return c.handleAftermarketDeviceClaimedEvent(ctx, &data)
		case AftermarketDevicePaired:
			return c.handleAftermarketDevicePairedEvent(ctx, &data)
		case AftermarketDeviceUnpaired:
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

	case DCNRegistryAddr:
		switch eventName {
		case NewNode:
			return c.handleNewDcnNode(ctx, &data)
		case NewExpiration:
			return c.handleNewDCNExpiration(ctx, &data)
		}
	case DCNResolverAddr:
		switch eventName {
		case NameChanged:
			return c.handleNameChanged(ctx, &data)
		case VehicleIdChanged:
			return c.handleVehicleIdChanged(ctx, &data)
		}
	}

	c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")

	return nil
}

func (c *ContractsEventsConsumer) handleManufacturerNodeMintedEvent(ctx context.Context, e *ContractEventData) error {
	var args ManufacturerNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	mfr := models.Manufacturer{
		ID:       int(args.TokenID.Int64()),
		Name:     args.Name,
		Owner:    args.Owner.Bytes(),
		MintedAt: e.Block.Time,
	}

	return mfr.Upsert(ctx, c.dbs.DBS().Writer, false, []string{models.ManufacturerColumns.ID}, boil.None(), boil.Infer())
}

func (c *ContractsEventsConsumer) handleVehicleNodeMintedEvent(ctx context.Context, e *ContractEventData) error {
	var args VehicleNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	v := models.Vehicle{
		ID:             int(args.TokenID.Int64()),
		OwnerAddress:   args.Owner.Bytes(),
		MintedAt:       e.Block.Time,
		ManufacturerID: null.IntFrom(int(args.ManufacturerNode.Int64())),
	}

	cols := models.VehicleColumns

	return v.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		false,
		[]string{cols.ID},
		boil.None(),
		boil.Whitelist(cols.ID, cols.OwnerAddress, cols.MintedAt),
	)
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceMintedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDeviceNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:             int(args.TokenID.Int64()),
		Address:        args.AftermarketDeviceAddress.Bytes(),
		Owner:          args.Owner.Bytes(),
		MintedAt:       e.Block.Time,
		Beneficiary:    args.Owner.Bytes(),
		ManufacturerID: null.IntFrom(int(args.ManufacturerID.Int64())),
	}

	cols := models.AftermarketDeviceColumns

	return ad.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		false,
		[]string{cols.ID},
		boil.None(),
		boil.Whitelist(cols.ID, cols.Address, cols.Owner, cols.MintedAt, cols.Beneficiary),
	)
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
		switch args.Attribute {
		case "Make":
			veh.Make = null.StringFrom(args.Info)
			_, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.Make))
			return err
		case "Model":
			veh.Model = null.StringFrom(args.Info)
			_, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.Model))
			return err
		case "Year":
			year, err := strconv.Atoi(args.Info)
			if err != nil {
				return fmt.Errorf("couldn't parse year string %q: %w", args.Info, err)
			}
			veh.Year = null.IntFrom(year)
			_, err = veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.Year))
			return err
		}
		return nil
	case "DefinitionURI":
		res, err := c.httpClient.Get(args.Info)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("device definition URI returned status code %d", res.StatusCode)
		}

		var ddf DeviceDefinition
		if err := json.NewDecoder(res.Body).Decode(&ddf); err != nil {
			return fmt.Errorf("couldn't parse device definition response: %w", err)
		}

		veh.Make = null.StringFrom(ddf.Type.Make)
		veh.Model = null.StringFrom(ddf.Type.Model)
		veh.Year = null.IntFrom(ddf.Type.Year)
		veh.DefinitionURI = null.StringFrom(args.Info)

		cols := models.VehicleColumns
		_, err = veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(cols.DefinitionURI, cols.Make, cols.Model, cols.Year))

		return err
	default:
		return fmt.Errorf("unrecognized vehicle attribute %q", args.Attribute)
	}
}

func (c *ContractsEventsConsumer) handleVehicleTransferEvent(ctx context.Context, e *ContractEventData) error {
	logger := c.log.With().Str("EventName", Transfer.String()).Logger()

	var args TransferData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	vehicle := models.Vehicle{
		ID:           int(args.TokenID.Int64()),
		OwnerAddress: args.To.Bytes(),
		MintedAt:     e.Block.Time,
	}

	// Handle this with VehicleNodeMinted.
	if args.From == zeroAddress {
		return nil
	}

	if args.To == zeroAddress {
		// Will cascade to the privileges.
		_, err := vehicle.Delete(ctx, c.dbs.DBS().Writer)
		return err
	}

	_, err := vehicle.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.OwnerAddress))
	if err != nil {
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

func (c *ContractsEventsConsumer) handleAftermarketDeviceClaimedEvent(ctx context.Context, e *ContractEventData) error {
	var args AftermarketDeviceClaimedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:        int(args.AftermarketDeviceNode.Int64()),
		ClaimedAt: null.TimeFrom(e.Block.Time),
	}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.ClaimedAt))
	return err
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
	var args TransferData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	if args.From == zeroAddress {
		// We handle mints via AftermarketDeviceNodeMinted.
		return nil
	}

	ad := models.AftermarketDevice{
		ID:          int(args.TokenID.Int64()),
		Owner:       args.To.Bytes(),
		Beneficiary: args.To.Bytes(),
	}

	_, err := ad.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.Beneficiary),
	)

	return err
}

func (c *ContractsEventsConsumer) handleBeneficiarySetEvent(ctx context.Context, e *ContractEventData) error {
	var args BeneficiarySetData
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

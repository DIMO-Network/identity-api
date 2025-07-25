package services

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/dbtypes"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/services/connection"
	cmodels "github.com/DIMO-Network/identity-api/internal/services/models"
	"github.com/DIMO-Network/identity-api/internal/services/staking"
	"github.com/DIMO-Network/identity-api/internal/services/storagenode"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/strings"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

type ContractsEventsConsumer struct {
	dbs                db.Store
	log                *zerolog.Logger
	settings           *config.Settings
	httpClient         *http.Client
	stakingHandler     *staking.Handler
	connsHandler       *connection.Handler
	storageNodeHandler *storagenode.Handler
}

type EventName string

var zeroAddress common.Address

const (
	// All NFTs.
	Transfer            EventName = "Transfer"
	PrivilegeSet        EventName = "PrivilegeSet"
	BeneficiarySetEvent EventName = "BeneficiarySet"

	// SACD.
	PermissionsSetEvent EventName = "PermissionsSet"

	// Manufacturers.
	ManufacturerNodeMinted       EventName = "ManufacturerNodeMinted"
	DeviceDefinitionTableCreated EventName = "DeviceDefinitionTableCreated"
	ManufacturerTableSet         EventName = "ManufacturerTableSet"

	// Aftermarket devices.
	AftermarketDeviceNodeMinted   EventName = "AftermarketDeviceNodeMinted"
	AftermarketDeviceAttributeSet EventName = "AftermarketDeviceAttributeSet"
	AftermarketDeviceClaimed      EventName = "AftermarketDeviceClaimed"
	AftermarketDevicePaired       EventName = "AftermarketDevicePaired"
	AftermarketDeviceUnpaired     EventName = "AftermarketDeviceUnpaired"
	AftermarketDeviceAddressReset EventName = "AftermarketDeviceAddressReset"

	// Vehicles.
	VehicleNodeMinted                     EventName = "VehicleNodeMinted"
	VehicleAttributeSet                   EventName = "VehicleAttributeSet"
	VehicleNodeMintedWithDeviceDefinition EventName = "VehicleNodeMintedWithDeviceDefinition"
	DeviceDefinitionIdSet                 EventName = "DeviceDefinitionIdSet"
	VehicleStorageNodeIdSet               EventName = "VehicleStorageNodeIdSet"

	// Synthetic devices.
	SyntheticDeviceNodeMinted EventName = "SyntheticDeviceNodeMinted"
	SyntheticDeviceNodeBurned EventName = "SyntheticDeviceNodeBurned"

	// DCNs.
	NewNode          EventName = "NewNode"
	NewExpiration    EventName = "NewExpiration"
	NameChanged      EventName = "NameChanged"
	VehicleIdChanged EventName = "VehicleIdChanged"

	// Rewards.
	TokensTransferredForDevice           EventName = "TokensTransferredForDevice"
	TokensTransferredForConnectionStreak EventName = "TokensTransferredForConnectionStreak"

	// Developer licenses.
	Issued              EventName = "Issued"
	RedirectUriEnabled  EventName = "RedirectUriEnabled"
	RedirectUriDisabled EventName = "RedirectUriDisabled"
	SignerEnabled       EventName = "SignerEnabled"
	SignerDisabled      EventName = "SignerDisabled"
	LicenseAliasSet     EventName = "LicenseAliasSet"
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
		stakingHandler: &staking.Handler{
			DBS: dbs,
		},
		connsHandler:       &connection.Handler{DBS: dbs, Logger: log},
		storageNodeHandler: &storagenode.Handler{DBS: dbs, Logger: log},
	}
}

func (c *ContractsEventsConsumer) Process(ctx context.Context, event *cloudevent.RawEvent) error {
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
	RewardsContractAddr := common.HexToAddress(c.settings.RewardsContractAddr)
	sacdAddr := common.HexToAddress(c.settings.SACDAddress)
	devLicenseAddr := common.HexToAddress(c.settings.DevLicenseAddr)
	stakingAddr := common.HexToAddress(c.settings.StakingAddr)
	connAddr := common.HexToAddress(c.settings.ConnectionAddr)
	manufacturerAddr := common.HexToAddress(c.settings.ManufacturerNFTAddr)
	storageNodeAddr := common.HexToAddress(c.settings.StorageNodeAddr)

	var data cmodels.ContractEventData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	eventName := EventName(data.EventName)

	c.log.Debug().Str("Event", string(eventName)).Str("Contract", data.Contract.Hex()).Msg("Event Received")

	switch data.Contract {
	case registryAddr:
		switch eventName {
		case ManufacturerNodeMinted:
			return c.handleManufacturerNodeMintedEvent(ctx, &data)
		case DeviceDefinitionTableCreated:
			return c.handleDeviceDefinitionTableCreated(ctx, &data)
		case ManufacturerTableSet:
			return c.handleManufacturerTableSet(ctx, &data)

		case VehicleNodeMinted:
			return c.handleVehicleNodeMintedEvent(ctx, &data)
		case VehicleNodeMintedWithDeviceDefinition:
			return c.handleVehicleNodeMintedWithDeviceDefinitionEvent(ctx, &data)
		case VehicleAttributeSet:
			return c.handleVehicleAttributeSetEvent(ctx, &data)
		case DeviceDefinitionIdSet:
			return c.handleDeviceDefinitionIdSet(ctx, &data)
		case VehicleStorageNodeIdSet:
			return c.handleNodeIdSetForVehicleID(ctx, &data)

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
		case AftermarketDeviceAddressReset:
			return c.handleAftermarketDeviceAddressResetEvent(ctx, &data)

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
	case sacdAddr:
		switch eventName {
		case PermissionsSetEvent:
			return c.handlePermissionsSetEvent(ctx, &data)
		}
	case aftermarketDeviceAddr:
		switch eventName {
		case Transfer:
			return c.handleAftermarketDeviceTransferredEvent(ctx, &data)
		}

	case manufacturerAddr:
		switch eventName {
		case Transfer:
			return c.handleManufacturerTransferEvent(ctx, &data)
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
	case RewardsContractAddr:
		switch eventName {
		case TokensTransferredForDevice:
			return c.handleTokensTransferredForDevice(ctx, &data)
		case TokensTransferredForConnectionStreak:
			return c.handleTokensTransferredForConnectionStreak(ctx, &data)
		}
	case devLicenseAddr:
		switch eventName {
		case Issued:
			return c.handleDevLicenseIssued(ctx, &data)
		case LicenseAliasSet:
			return c.handleDevLicenseAlias(ctx, &data)
		case RedirectUriEnabled:
			return c.handleRedirectEnabled(ctx, &data)
		case RedirectUriDisabled:
			return c.handleRedirectDisabled(ctx, &data)
		case SignerEnabled:
			return c.handleSignerEnabled(ctx, &data)
		case SignerDisabled:
			return c.handleSignerDisabled(ctx, &data)
		}
	case stakingAddr:
		return c.stakingHandler.HandleEvent(ctx, &data)
	case connAddr:
		return c.connsHandler.Handle(ctx, &data)
	case storageNodeAddr:
		return c.storageNodeHandler.Handle(ctx, &data)
	}

	c.log.Debug().Str("event", data.EventName).Msg("Handler not provided for event.")

	return nil
}

func (c *ContractsEventsConsumer) handleManufacturerNodeMintedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args ManufacturerNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	mfr := models.Manufacturer{
		ID:       int(args.TokenID.Int64()),
		Name:     args.Name,
		Owner:    args.Owner.Bytes(),
		MintedAt: e.Block.Time,
		Slug:     strings.SlugString(args.Name), // Better hope uniqueness is never a problem!
	}

	return mfr.Upsert(ctx, c.dbs.DBS().Writer, false, []string{models.ManufacturerColumns.ID}, boil.None(), boil.Infer())
}

func (c *ContractsEventsConsumer) handleDeviceDefinitionTableCreated(ctx context.Context, e *cmodels.ContractEventData) error {
	var args DeviceDefinitionTableCreatedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	mfr := models.Manufacturer{
		ID:      int(args.ManufacturerId.Int64()),
		TableID: null.IntFrom(int(args.TableId.Int64())),
	}

	_, err := mfr.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.ManufacturerColumns.TableID))
	return err
}

func (c *ContractsEventsConsumer) handleManufacturerTableSet(ctx context.Context, e *cmodels.ContractEventData) error {
	var args ManufacturerTableSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	mfr := models.Manufacturer{
		ID:      int(args.ManufacturerId.Int64()),
		TableID: null.IntFrom(int(args.TableId.Int64())),
	}

	_, err := mfr.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.ManufacturerColumns.TableID))
	return err
}

func (c *ContractsEventsConsumer) handleVehicleNodeMintedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args VehicleNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	v := models.Vehicle{
		ID:             int(args.TokenID.Int64()),
		OwnerAddress:   args.Owner.Bytes(),
		MintedAt:       e.Block.Time,
		ManufacturerID: int(args.ManufacturerNode.Int64()),
	}

	cols := models.VehicleColumns

	return v.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		false,
		[]string{cols.ID},
		boil.None(),
		boil.Whitelist(cols.ID, cols.OwnerAddress, cols.MintedAt, cols.ManufacturerID),
	)
}

func (c *ContractsEventsConsumer) handleVehicleNodeMintedWithDeviceDefinitionEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args VehicleNodeMintedWithDeviceDefinitionData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	v := models.Vehicle{
		ID:                 int(args.VehicleId.Int64()),
		OwnerAddress:       args.Owner.Bytes(),
		DeviceDefinitionID: null.StringFrom(args.DeviceDefinitionID),
		MintedAt:           e.Block.Time,
		ManufacturerID:     int(args.ManufacturerId.Int64()),
	}

	cols := models.VehicleColumns

	return v.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		false,
		[]string{cols.ID},
		boil.None(),
		boil.Whitelist(cols.ID, cols.OwnerAddress, cols.MintedAt, cols.ManufacturerID, cols.DeviceDefinitionID),
	)
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceMintedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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
		ManufacturerID: int(args.ManufacturerID.Int64()),
	}

	cols := models.AftermarketDeviceColumns

	return ad.Upsert(
		ctx,
		c.dbs.DBS().Writer,
		false,
		[]string{cols.ID},
		boil.None(),
		boil.Whitelist(cols.ID, cols.Address, cols.Owner, cols.MintedAt, cols.Beneficiary, cols.ManufacturerID),
	)
}

func (c *ContractsEventsConsumer) handleVehicleAttributeSetEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args VehicleAttributeSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	veh := models.Vehicle{
		ID: int(args.TokenID.Int64()),
	}

	switch args.Attribute {
	case "Make":
		var make null.String
		if args.Info != "" {
			make = null.StringFrom(args.Info)
		}
		veh.Make = make
		_, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.Make))
		return err
	case "Model":
		var model null.String
		if args.Info != "" {
			model = null.StringFrom(args.Info)
		}
		veh.Model = model
		_, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.Model))
		return err
	case "Year":
		var year null.Int
		if args.Info != "" {
			yr, err := strconv.Atoi(args.Info)
			if err != nil {
				return fmt.Errorf("couldn't parse year string %q: %w", args.Info, err)
			}
			year = null.IntFrom(yr)
		}
		veh.Year = year
		_, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.Year))
		return err
	case "ImageURI":
		var imageURI null.String
		if args.Info != "" {
			_, err := url.ParseRequestURI(args.Info)
			if err != nil {
				return fmt.Errorf("couldn't parse image URI string %q: %w", args.Info, err)
			}
			imageURI = null.StringFrom(args.Info)
		}
		veh.ImageURI = imageURI
		_, err := veh.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.ImageURI))
		return err
	case "DefinitionURI", "DataURI":
		// We never ended up using these.
		return nil
	default:
		return fmt.Errorf("unrecognized vehicle attribute %q", args.Attribute)
	}
}

func (c *ContractsEventsConsumer) handleDeviceDefinitionIdSet(ctx context.Context, e *cmodels.ContractEventData) error {
	logger := c.log.With().Str("EventName", Transfer.String()).Logger()

	var args DeviceDefinitionIdSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	vehicle := models.Vehicle{
		ID:                 int(args.VehicleId.Int64()),
		DeviceDefinitionID: null.StringFrom(args.DDID),
	}

	// TODO(elffjs): Should we try to update the MMY fields using Tableland?
	// TODO(elffjs): Maybe it's interesting if the update count is zero?
	_, err := vehicle.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.DeviceDefinitionID))
	if err != nil {
		return err
	}

	logger.Info().Int64("vehicleId", args.VehicleId.Int64()).Msgf("Vehicle definition updated to %s.", args.DDID)

	return nil
}

func (c *ContractsEventsConsumer) handleNodeIdSetForVehicleID(ctx context.Context, e *cmodels.ContractEventData) error {
	logger := c.log.With().Str("EventName", Transfer.String()).Logger()

	var args VehicleStorageNodeIdSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	anchorID, err := helpers.ConvertTokenIDToID(args.StorageNodeId)
	if err != nil {
		return err
	}

	vehicle := models.Vehicle{
		ID:            int(args.VehicleId.Int64()),
		StorageNodeID: null.BytesFrom(anchorID),
	}

	_, err = vehicle.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.VehicleColumns.StorageNodeID))
	if err != nil {
		return err
	}

	// Don't want to incur the cost of a SELECT, and I don't know how to do RETURNING with SQLBoiler.
	// This at least is enough to go on.
	logger.Info().Int64("vehicleId", args.VehicleId.Int64()).Msgf("Vehicle storage node set to %d.", args.StorageNodeId)

	return nil
}

func (c *ContractsEventsConsumer) handleVehicleTransferEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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
		return fmt.Errorf("failed to delete associated privileges: %w", err)
	}

	if _, err := models.VehicleSacds(models.VehicleSacdWhere.VehicleID.EQ(int(args.TokenID.Int64()))).DeleteAll(ctx, c.dbs.DBS().Writer); err != nil {
		return fmt.Errorf("failed to delete associated SACDs: %w", err)
	}

	logger.Info().Str("TokenID", args.TokenID.String()).Msg("Event processed successfuly")

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceAttributeSetEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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
	case "DevEUI":
		ad.DevEui = null.StringFrom(args.Info)
		if _, err := ad.Update(
			ctx,
			c.dbs.DBS().Writer,
			boil.Whitelist(models.AftermarketDeviceColumns.DevEui)); err != nil {
			return err
		}
	case "HardwareRevision":
		ad.HardwareRevision = null.StringFrom(args.Info)
		if _, err := ad.Update(
			ctx,
			c.dbs.DBS().Writer,
			boil.Whitelist(models.AftermarketDeviceColumns.HardwareRevision)); err != nil {
			return err
		}
	}

	return nil
}

func (c *ContractsEventsConsumer) handlePermissionsSetEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	logger := c.log.With().Str("EventName", PermissionsSetEvent.String()).Logger()

	var args PermissionsSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return fmt.Errorf("error unmarshaling PermissionsSet inputs: %w", err)
	}

	if args.Asset != common.HexToAddress(c.settings.VehicleNFTAddr) {
		logger.Warn().Msgf("SACD set for non-vehicle asset %s.", args.Asset.Hex())
		return nil
	}

	sacd := models.VehicleSacd{
		VehicleID:   int(args.TokenId.Int64()),
		Grantee:     args.Grantee.Bytes(),
		Permissions: args.Permissions.Text(2),
		Source:      args.Source,
		CreatedAt:   e.Block.Time,
		ExpiresAt:   time.Unix(args.Expiration.Int64(), 0),
	}

	if err := sacd.Upsert(ctx, c.dbs.DBS().Writer, true,
		[]string{
			models.VehicleSacdColumns.VehicleID,
			models.VehicleSacdColumns.Grantee,
		},
		boil.Whitelist(models.VehicleSacdColumns.Permissions, models.VehicleSacdColumns.Source, models.VehicleSacdColumns.CreatedAt, models.PrivilegeColumns.ExpiresAt),
		boil.Infer()); err != nil {
		return fmt.Errorf("error upserting vehicle SACD: %w", err)
	}

	logger.Info().
		Int64("vehicleId", args.TokenId.Int64()).
		Str("grantee", args.Grantee.Hex()).
		Msg("Vehicle SACD processed.")

	return nil
}

func (c *ContractsEventsConsumer) handlePrivilegeSetEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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

func (c *ContractsEventsConsumer) handleAftermarketDeviceClaimedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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

func (c *ContractsEventsConsumer) handleAftermarketDevicePairedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{
		ID:        int(args.AftermarketDeviceNode.Int64()),
		VehicleID: null.IntFrom(int(args.VehicleNode.Int64())),
		PairedAt:  null.TimeFrom(e.Block.Time),
	}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleID, models.AftermarketDeviceColumns.PairedAt))
	if err != nil {
		return err
	}

	c.log.Info().Int64("vehicleId", args.VehicleNode.Int64()).Int64("aftermarketId", args.AftermarketDeviceNode.Int64()).Msg("Aftermarket device paired.")

	return nil
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceUnpairedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args AftermarketDevicePairData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ad := models.AftermarketDevice{ID: int(args.AftermarketDeviceNode.Int64())}

	_, err := ad.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.VehicleID, models.AftermarketDeviceColumns.PairedAt))
	return err
}

func (c *ContractsEventsConsumer) handleManufacturerTransferEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args TransferData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	if args.From == zeroAddress {
		// We handle mints via ManufacturerNodeMinted.
		return nil
	}

	mfr := models.Manufacturer{
		ID:    int(args.TokenID.Int64()),
		Owner: args.To.Bytes(),
	}

	if args.To == zeroAddress {
		// Must be a burn.
		_, err := mfr.Delete(ctx, c.dbs.DBS().Writer)
		return err
	}

	_, err := mfr.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.ManufacturerColumns.Owner),
	)

	return err
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceTransferredEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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

	if args.To == zeroAddress {
		// Must be a burn.
		_, err := ad.Delete(ctx, c.dbs.DBS().Writer)
		return err
	}

	_, err := ad.Update(
		ctx,
		c.dbs.DBS().Writer,
		boil.Whitelist(models.AftermarketDeviceColumns.Owner, models.AftermarketDeviceColumns.Beneficiary),
	)

	return err
}

func (c *ContractsEventsConsumer) handleBeneficiarySetEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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

func takeFirst(x, y *big.Int) *big.Int {
	if x != nil {
		return x
	}
	return y
}

func (c *ContractsEventsConsumer) handleSyntheticDeviceNodeMintedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args SyntheticDeviceNodeMintedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	var integrationID int
	var connectionID null.Bytes

	rightID := takeFirst(args.ConnectionID, args.IntegrationNode) // Both fields being nil would be very bad.

	if rightID.IsInt64() {
		integrationID = int(rightID.Int64())
	} else {
		id, err := helpers.ConvertTokenIDToID(rightID)
		if err != nil {
			return err
		}
		connectionID = null.BytesFrom(id)
	}

	sd := models.SyntheticDevice{
		ID:            int(args.SyntheticDeviceNode.Int64()),
		IntegrationID: integrationID,
		VehicleID:     int(args.VehicleNode.Int64()),
		DeviceAddress: args.SyntheticDeviceAddress.Bytes(),
		MintedAt:      e.Block.Time,
		ConnectionID:  connectionID,
	}

	return sd.Insert(ctx, c.dbs.DBS().Writer, boil.Infer())
}

func (c *ContractsEventsConsumer) handleSyntheticDeviceNodeBurnedEvent(ctx context.Context, e *cmodels.ContractEventData) error {
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

func (c *ContractsEventsConsumer) handleTokensTransferredForDevice(ctx context.Context, e *cmodels.ContractEventData) error {
	var args TokensTransferredForDeviceData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	reward := models.Reward{
		IssuanceWeek:      int(args.Week.Int64()),
		VehicleID:         int(args.VehicleNodeID.Int64()),
		ReceivedByAddress: null.BytesFrom(args.User.Bytes()),
		EarnedAt:          e.Block.Time,
	}

	cols := models.RewardColumns

	if common.HexToAddress(c.settings.AftermarketDeviceAddr) == args.DeviceNftProxy {
		reward.AftermarketTokenID = null.IntFrom(int(args.DeviceNode.Int64()))
		reward.AftermarketEarnings = dbtypes.IntToDecimal(args.Amount)
		return reward.Upsert(ctx, c.dbs.DBS().Writer, true,
			[]string{cols.IssuanceWeek, cols.VehicleID},
			boil.Whitelist(cols.AftermarketEarnings, cols.AftermarketTokenID),
			boil.Infer())
	} else if common.HexToAddress(c.settings.SyntheticDeviceAddr) == args.DeviceNftProxy {
		reward.SyntheticTokenID = null.IntFrom(int(args.DeviceNode.Int64()))
		reward.SyntheticEarnings = dbtypes.IntToDecimal(args.Amount)

		return reward.Upsert(ctx, c.dbs.DBS().Writer, true,
			[]string{cols.IssuanceWeek, cols.VehicleID},
			boil.Whitelist(cols.SyntheticEarnings, cols.SyntheticTokenID),
			boil.Infer())
	}

	return nil
}

func (c *ContractsEventsConsumer) handleTokensTransferredForConnectionStreak(ctx context.Context, e *cmodels.ContractEventData) error {
	var args TokensTransferredForConnectionStreakData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	reward := models.Reward{
		IssuanceWeek:     int(args.Week.Int64()),
		VehicleID:        int(args.VehicleNodeID.Int64()),
		ConnectionStreak: null.IntFrom(int(args.ConnectionStreak.Int64())),
		StreakEarnings:   dbtypes.IntToDecimal(args.Amount),
	}

	cols := models.RewardColumns

	_, err := reward.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(cols.StreakEarnings, cols.ConnectionStreak))

	return err
}

func (c *ContractsEventsConsumer) handleAftermarketDeviceAddressResetEvent(ctx context.Context, e *cmodels.ContractEventData) error {
	var args AftermarketDeviceAddressResetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	amd, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.ID.EQ(int(args.TokenId.Int64())),
	).One(ctx, c.dbs.DBS().Reader)
	if err != nil {
		return err
	}

	amd.Address = args.AftermarketDeviceAddress.Bytes()
	_, err = amd.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.AftermarketDeviceColumns.Address))
	return err
}

func (c *ContractsEventsConsumer) handleDevLicenseIssued(ctx context.Context, e *cmodels.ContractEventData) error {
	var args IssuedData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dl := models.DeveloperLicense{
		ID:       int(args.TokenID.Int64()),
		Owner:    args.Owner.Bytes(),
		ClientID: args.ClientID.Bytes(),
		MintedAt: e.Block.Time,
	}

	err := dl.Upsert(ctx, c.dbs.DBS().Writer, false, []string{models.DeveloperLicenseColumns.ID}, boil.Blacklist(), boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleDevLicenseAlias(ctx context.Context, e *cmodels.ContractEventData) error {
	var args LicenseAliasSetData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	dlID := int(args.TokenID.Int64())

	alias := args.LicenseAlias

	var dbAlias null.String

	if alias != "" {
		dbAlias = null.StringFrom(alias)
	}

	dl := models.DeveloperLicense{
		ID:    dlID,
		Alias: dbAlias,
	}

	c.log.Info().Int("developerLicenseId", dlID).Msgf("Developer license alias set to %q.", alias)

	_, err := dl.Update(ctx, c.dbs.DBS().Writer, boil.Whitelist(models.DeveloperLicenseColumns.Alias))
	if err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleRedirectEnabled(ctx context.Context, e *cmodels.ContractEventData) error {
	var args RedirectUriEnabledData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ru := models.RedirectURI{
		DeveloperLicenseID: int(args.TokenID.Int64()),
		URI:                args.URI,
		EnabledAt:          e.Block.Time,
	}

	err := ru.Upsert(ctx, c.dbs.DBS().Writer, false, []string{models.RedirectURIColumns.DeveloperLicenseID, models.RedirectURIColumns.URI}, boil.Blacklist(), boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleRedirectDisabled(ctx context.Context, e *cmodels.ContractEventData) error {
	var args RedirectUriDisabledData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	ru := models.RedirectURI{
		DeveloperLicenseID: int(args.TokenID.Int64()),
		URI:                args.URI,
	}

	_, err := ru.Delete(ctx, c.dbs.DBS().Writer)
	return err
}

func (c *ContractsEventsConsumer) handleSignerEnabled(ctx context.Context, e *cmodels.ContractEventData) error {
	var args SignerEnabledData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	s := models.Signer{
		DeveloperLicenseID: int(args.TokenID.Int64()),
		Signer:             args.Signer.Bytes(),
		EnabledAt:          e.Block.Time,
	}

	err := s.Upsert(ctx, c.dbs.DBS().Writer, false, []string{models.SignerColumns.DeveloperLicenseID, models.SignerColumns.Signer}, boil.Blacklist(), boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (c *ContractsEventsConsumer) handleSignerDisabled(ctx context.Context, e *cmodels.ContractEventData) error {
	var args SignerDisabledData
	if err := json.Unmarshal(e.Arguments, &args); err != nil {
		return err
	}

	s := models.Signer{
		DeveloperLicenseID: int(args.TokenID.Int64()),
		Signer:             args.Signer.Bytes(),
	}

	_, err := s.Delete(ctx, c.dbs.DBS().Writer)
	return err
}

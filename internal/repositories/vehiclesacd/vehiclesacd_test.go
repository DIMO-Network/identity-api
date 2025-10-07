package vehiclesacd

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/identity-api/internal/config"
	"github.com/DIMO-Network/identity-api/internal/helpers"
	"github.com/DIMO-Network/identity-api/internal/repositories/base"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type VehiclesSacdRepoTestSuite struct {
	suite.Suite
	ctx       context.Context
	pdb       db.Store
	container *postgres.PostgresContainer
	repo      *Repository
	settings  config.Settings
}

func (s *VehiclesSacdRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = helpers.StartContainerDatabase(s.ctx, s.T(), "../../../migrations")

	s.settings = config.Settings{
		DIMORegistryAddr:    "0x4de1bcf2b7e851e31216fc07989caa902a604784",
		DIMORegistryChainID: 80001,
	}
	logger := zerolog.Nop()
	s.repo = &Repository{base.NewRepository(s.pdb, s.settings, &logger)}
}

// TearDownTest after each test truncate tables
func (s *VehiclesSacdRepoTestSuite) TearDownTest() {
	s.Require().NoError(s.container.Restore(s.ctx))
}

// TearDownSuite cleanup at end by terminating container
func (s *VehiclesSacdRepoTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())

	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

// Test Runner
func TestVehiclesPrivilegesRepoTestSuite(t *testing.T) {
	suite.Run(t, new(VehiclesSacdRepoTestSuite))
}

func (s *VehiclesSacdRepoTestSuite) TestSacdToAPIResponse_WithoutTemplate() {
	sacd := &models.VehicleSacd{
		VehicleID:   1,
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	result, err := sacdToAPIResponse(sacd)
	s.NoError(err)
	s.NotNil(result)
	s.Nil(result.Template)
	s.Equal("0xa", result.Permissions) // 1010 binary = a hex
}

func (s *VehiclesSacdRepoTestSuite) TestSacdToAPIResponse_WithTemplate() {
	templateIdStr := "39432737238797479986393736422353506685818102154178527542856781838379111614015"
	templateId := new(big.Int)
	templateId.SetString(templateIdStr, 10)
	tokenIdBytes, err := helpers.ConvertTokenIDToID(templateId)
	assert.NoError(s.T(), err)

	creator := common.HexToAddress("0x1111111111111111111111111111111111111111")
	asset := common.HexToAddress("0x2222222222222222222222222222222222222222")

	template := &models.Template{
		ID:          tokenIdBytes,
		Creator:     creator.Bytes(),
		Asset:       asset.Bytes(),
		Permissions: "1111",
		Cid:         "QmTestCID123",
		CreatedAt:   time.Now(),
	}

	sacd := &models.VehicleSacd{
		VehicleID:   1,
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	// Manually set the relationship for testing
	sacd.R = sacd.R.NewStruct()
	sacd.R.Template = template

	result, err := sacdToAPIResponse(sacd)
	s.NoError(err)
	s.NotNil(result)
	s.NotNil(result.Template)

	s.Equal(new(big.Int).SetBytes(tokenIdBytes), result.Template.TokenID)
	s.Equal(creator, result.Template.Creator)
	s.Equal(asset, result.Template.Asset)
	s.Equal("1111", result.Template.Permissions)
	s.Equal("QmTestCID123", result.Template.Cid)
	s.Equal(template.CreatedAt, result.Template.CreatedAt)
}

func (s *VehiclesSacdRepoTestSuite) TestSacdToAPIResponse_WithRButNoTemplate() {
	sacd := &models.VehicleSacd{
		VehicleID:   1,
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	// Manually set the relationship for testing with nil template
	sacd.R = sacd.R.NewStruct()
	sacd.R.Template = nil

	result, err := sacdToAPIResponse(sacd)
	s.NoError(err)
	s.NotNil(result)
	s.Nil(result.Template)
}

func (s *VehiclesSacdRepoTestSuite) TestSacdToAPIResponse_InvalidPermissions() {
	sacd := &models.VehicleSacd{
		VehicleID:   1,
		Grantee:     common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes(),
		Permissions: "invalid-binary",
		Source:      "test-source",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	result, err := sacdToAPIResponse(sacd)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "couldn't parse permission string")
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_Success() {
	_, granteeWallet, err := helpers.GenerateWallet()
	s.NoError(err)
	_, ownerWallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicle := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   ownerWallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Corolla"),
		Year:           null.IntFrom(2020),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	templateIdStr := "39432737238797479986393736422353506685818102154178527542856781838379111614015"
	templateId := new(big.Int)
	templateId.SetString(templateIdStr, 10)
	tokenIdBytes, err := helpers.ConvertTokenIDToID(templateId)
	assert.NoError(s.T(), err)

	template := models.Template{
		ID:          tokenIdBytes,
		Creator:     common.HexToAddress("0x1111111111111111111111111111111111111111").Bytes(),
		Asset:       common.HexToAddress("0x2222222222222222222222222222222222222222").Bytes(),
		Permissions: "1111",
		Cid:         "QmTestCID123",
		CreatedAt:   currTime,
	}

	if err := template.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	expiresAt := currTime.Add(time.Hour).UTC().Truncate(time.Second)
	sacd := models.VehicleSacd{
		VehicleID:   1,
		Grantee:     granteeWallet.Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		TemplateID:  null.BytesFrom(tokenIdBytes),
		CreatedAt:   currTime,
		ExpiresAt:   expiresAt,
	}

	if err := sacd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	first := 10
	res, err := s.repo.GetSacdsForVehicle(s.ctx, 1, &first, nil, nil, nil)
	s.NoError(err)

	s.NotNil(res)
	s.Equal(1, res.TotalCount)
	s.Len(res.Edges, 1)
	s.Len(res.Nodes, 1)

	returnedSacd := res.Nodes[0]
	s.Equal(*granteeWallet, returnedSacd.Grantee)
	s.Equal("0xa", returnedSacd.Permissions) // 1010 binary = a hex
	s.Equal("test-source", returnedSacd.Source)
	s.Equal(currTime, returnedSacd.CreatedAt)
	s.Equal(expiresAt, returnedSacd.ExpiresAt)

	// Verify template relationship
	s.NotNil(returnedSacd.Template)
	s.Equal(new(big.Int).SetBytes(tokenIdBytes), returnedSacd.Template.TokenID)
	s.Equal(common.HexToAddress("0x1111111111111111111111111111111111111111"), returnedSacd.Template.Creator)
	s.Equal(common.HexToAddress("0x2222222222222222222222222222222222222222"), returnedSacd.Template.Asset)
	s.Equal("1111", returnedSacd.Template.Permissions)
	s.Equal("QmTestCID123", returnedSacd.Template.Cid)
	s.Equal(template.CreatedAt, returnedSacd.Template.CreatedAt)
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_NoSacds() {
	_, ownerWallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicle := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   ownerWallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Corolla"),
		Year:           null.IntFrom(2021),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	first := 10
	res, err := s.repo.GetSacdsForVehicle(s.ctx, 1, &first, nil, nil, nil)
	s.NoError(err)

	s.NotNil(res)
	s.Equal(0, res.TotalCount)
	s.Len(res.Edges, 0)
	s.Len(res.Nodes, 0)
	s.NotNil(res.PageInfo)
	s.Nil(res.PageInfo.StartCursor)
	s.Nil(res.PageInfo.EndCursor)
	s.False(res.PageInfo.HasNextPage)
	s.False(res.PageInfo.HasPreviousPage)
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_OnlyExpiredSacds() {
	_, granteeWallet, err := helpers.GenerateWallet()
	s.NoError(err)
	_, ownerWallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	// Create manufacturer
	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicle := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   ownerWallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Corolla"),
		Year:           null.IntFrom(2022),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	expiredAt := currTime.Add(-time.Hour).UTC().Truncate(time.Second)
	sacd := models.VehicleSacd{
		VehicleID:   1,
		Grantee:     granteeWallet.Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   currTime.Add(-2 * time.Hour),
		ExpiresAt:   expiredAt,
	}

	if err := sacd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	// Test GetSacdsForVehicle - should return empty because SACD is expired
	first := 10
	res, err := s.repo.GetSacdsForVehicle(s.ctx, 1, &first, nil, nil, nil)
	s.NoError(err)

	// Verify empty response (expired SACDs are filtered out)
	s.NotNil(res)
	s.Equal(0, res.TotalCount)
	s.Len(res.Edges, 0)
	s.Len(res.Nodes, 0)
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_Pagination_FirstAfter() {
	_, grantee1, err := helpers.GenerateWallet()
	s.NoError(err)
	_, grantee2, err := helpers.GenerateWallet()
	s.NoError(err)
	_, ownerWallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicle := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   ownerWallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Corolla"),
		Year:           null.IntFrom(2023),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	// Create multiple SACDs with different creation times to test ordering
	expiresAt := currTime.Add(time.Hour).UTC().Truncate(time.Second)

	sacds := []models.VehicleSacd{
		{
			VehicleID:   1,
			Grantee:     grantee1.Bytes(),
			Permissions: "1010",
			Source:      "test-source-1",
			CreatedAt:   currTime,
			ExpiresAt:   expiresAt,
		},
		{
			VehicleID:   1,
			Grantee:     grantee2.Bytes(),
			Permissions: "1100",
			Source:      "test-source-2",
			CreatedAt:   currTime.Add(-time.Minute),
			ExpiresAt:   expiresAt,
		},
	}

	for _, sacd := range sacds {
		if err := sacd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	limit := 1
	res, err := s.repo.GetSacdsForVehicle(s.ctx, 1, &limit, nil, nil, nil)
	s.NoError(err)

	// Should return the most recent SACD first (DESC order by created_at)
	s.NotNil(res)
	s.Equal(2, res.TotalCount)
	s.Len(res.Edges, 1)
	s.Len(res.Nodes, 1)
	s.True(res.PageInfo.HasNextPage)
	s.False(res.PageInfo.HasPreviousPage)
	s.NotNil(res.PageInfo.EndCursor)

	// Should be the first SACD (most recent)
	s.Equal(*grantee1, res.Nodes[0].Grantee)
	s.Equal("0xa", res.Nodes[0].Permissions) // 1010 binary = a hex

	// Test second page using after cursor
	res2, err := s.repo.GetSacdsForVehicle(s.ctx, 1, &limit, res.PageInfo.EndCursor, nil, nil)
	s.NoError(err)

	s.NotNil(res2)
	s.Equal(2, res2.TotalCount)
	s.Len(res2.Edges, 1)
	s.Len(res2.Nodes, 1)
	s.False(res2.PageInfo.HasNextPage)
	s.True(res2.PageInfo.HasPreviousPage)

	// Should be the second SACD (older)
	s.Equal(*grantee2, res2.Nodes[0].Grantee)
	s.Equal("0xc", res2.Nodes[0].Permissions) // 1100 binary = c hex
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_Pagination_LastBefore() {
	_, grantee1, err := helpers.GenerateWallet()
	s.NoError(err)
	_, grantee2, err := helpers.GenerateWallet()
	s.NoError(err)
	_, ownerWallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicle := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   ownerWallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Corolla"),
		Year:           null.IntFrom(2023),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	// Create multiple SACDs
	expiresAt := currTime.Add(time.Hour).UTC().Truncate(time.Second)

	sacds := []models.VehicleSacd{
		{
			VehicleID:   1,
			Grantee:     grantee1.Bytes(),
			Permissions: "1010",
			Source:      "test-source-1",
			CreatedAt:   currTime,
			ExpiresAt:   expiresAt,
		},
		{
			VehicleID:   1,
			Grantee:     grantee2.Bytes(),
			Permissions: "1100",
			Source:      "test-source-2",
			CreatedAt:   currTime.Add(-time.Minute),
			ExpiresAt:   expiresAt,
		},
	}

	for _, sacd := range sacds {
		if err := sacd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
			s.NoError(err)
		}
	}

	// Get all SACDs first to get a cursor
	firstAll := 10
	res, err := s.repo.GetSacdsForVehicle(s.ctx, 1, &firstAll, nil, nil, nil)
	s.NoError(err)
	s.Len(res.Nodes, 2)

	limit := 1
	res2, err := s.repo.GetSacdsForVehicle(s.ctx, 1, nil, nil, &limit, res.PageInfo.EndCursor)
	s.NoError(err)

	s.NotNil(res2)
	s.Equal(2, res2.TotalCount)
	s.Len(res2.Edges, 1)
	s.Len(res2.Nodes, 1)
	s.True(res2.PageInfo.HasNextPage)
	s.False(res2.PageInfo.HasPreviousPage)

	// Should return the first SACD (most recent)
	s.Equal(*grantee1, res2.Nodes[0].Grantee)
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_WithoutTemplate() {
	_, granteeWallet, err := helpers.GenerateWallet()
	s.NoError(err)
	_, ownerWallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicle := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   ownerWallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Corolla"),
		Year:           null.IntFrom(2023),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	// Create SACD without template
	expiresAt := currTime.Add(time.Hour).UTC().Truncate(time.Second)
	sacd := models.VehicleSacd{
		VehicleID:   1,
		Grantee:     granteeWallet.Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   currTime,
		ExpiresAt:   expiresAt,
	}

	if err := sacd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	first := 10
	res, err := s.repo.GetSacdsForVehicle(s.ctx, 1, &first, nil, nil, nil)
	s.NoError(err)

	s.NotNil(res)
	s.Equal(1, res.TotalCount)
	s.Len(res.Edges, 1)
	s.Len(res.Nodes, 1)

	// Verify SACD data without template
	returnedSacd := res.Nodes[0]
	s.Equal(*granteeWallet, returnedSacd.Grantee)
	s.Equal("0xa", returnedSacd.Permissions)
	s.Equal("test-source", returnedSacd.Source)
	s.Nil(returnedSacd.Template)
}

func (s *VehiclesSacdRepoTestSuite) TestCreateSacdResponse() {
	_, grantee1, err := helpers.GenerateWallet()
	s.NoError(err)
	_, grantee2, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)
	expiresAt := currTime.Add(time.Hour).UTC().Truncate(time.Second)

	sacds := models.VehicleSacdSlice{
		&models.VehicleSacd{
			VehicleID:   1,
			Grantee:     grantee1.Bytes(),
			Permissions: "1010",
			Source:      "test-source-1",
			CreatedAt:   currTime,
			ExpiresAt:   expiresAt,
		},
		&models.VehicleSacd{
			VehicleID:   1,
			Grantee:     grantee2.Bytes(),
			Permissions: "1100",
			Source:      "test-source-2",
			CreatedAt:   currTime.Add(-time.Minute),
			ExpiresAt:   expiresAt,
		},
	}

	pHelper := helpers.PaginationHelper[SacdCursor]{}

	result, err := s.repo.createSacdResponse(sacds, 2, true, false, pHelper)
	s.NoError(err)

	s.NotNil(result)
	s.Equal(2, result.TotalCount)
	s.Len(result.Edges, 2)
	s.Len(result.Nodes, 2)
	s.NotNil(result.PageInfo)
	s.True(result.PageInfo.HasNextPage)
	s.False(result.PageInfo.HasPreviousPage)
	s.NotNil(result.PageInfo.StartCursor)
	s.NotNil(result.PageInfo.EndCursor)

	// Verify first edge/node
	s.Equal(*grantee1, result.Edges[0].Node.Grantee)
	s.Equal("0xa", result.Edges[0].Node.Permissions)
	s.Equal("test-source-1", result.Edges[0].Node.Source)
	s.NotEmpty(result.Edges[0].Cursor)

	// Verify second edge/node
	s.Equal(*grantee2, result.Edges[1].Node.Grantee)
	s.Equal("0xc", result.Edges[1].Node.Permissions)
	s.Equal("test-source-2", result.Edges[1].Node.Source)
	s.NotEmpty(result.Edges[1].Cursor)

	// Verify nodes match edges
	s.Equal(result.Edges[0].Node, result.Nodes[0])
	s.Equal(result.Edges[1].Node, result.Nodes[1])
}

func (s *VehiclesSacdRepoTestSuite) TestCreateSacdResponse_EmptySacds() {
	pHelper := helpers.PaginationHelper[SacdCursor]{}

	result, err := s.repo.createSacdResponse(models.VehicleSacdSlice{}, 0, false, false, pHelper)
	s.NoError(err)

	s.NotNil(result)
	s.Equal(0, result.TotalCount)
	s.Len(result.Edges, 0)
	s.Len(result.Nodes, 0)
	s.NotNil(result.PageInfo)
	s.False(result.PageInfo.HasNextPage)
	s.False(result.PageInfo.HasPreviousPage)
	s.Nil(result.PageInfo.StartCursor)
	s.Nil(result.PageInfo.EndCursor)
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_InvalidCursor() {
	_, granteeWallet, err := helpers.GenerateWallet()
	s.NoError(err)
	_, ownerWallet, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	m := models.Manufacturer{
		ID:       131,
		Name:     "Toyota",
		Owner:    common.FromHex("0x46a3A41bd932244Dd08186e4c19F1a7E48cbcDf4"),
		MintedAt: time.Now(),
		Slug:     "toyota",
	}

	if err := m.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		assert.NoError(s.T(), err)
	}

	vehicle := models.Vehicle{
		ID:             1,
		ManufacturerID: 131,
		OwnerAddress:   ownerWallet.Bytes(),
		Make:           null.StringFrom("Toyota"),
		Model:          null.StringFrom("Corolla"),
		Year:           null.IntFrom(2023),
		MintedAt:       currTime,
	}

	if err := vehicle.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	expiresAt := currTime.Add(time.Hour).UTC().Truncate(time.Second)
	sacd := models.VehicleSacd{
		VehicleID:   1,
		Grantee:     granteeWallet.Bytes(),
		Permissions: "1010",
		Source:      "test-source",
		CreatedAt:   currTime,
		ExpiresAt:   expiresAt,
	}

	if err := sacd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer()); err != nil {
		s.NoError(err)
	}

	invalidCursor := "invalid-cursor!!!"
	first := 10
	_, err = s.repo.GetSacdsForVehicle(s.ctx, 1, &first, &invalidCursor, nil, nil)
	s.Error(err)
	s.Contains(err.Error(), "illegal base64 data")

	last := 10
	_, err = s.repo.GetSacdsForVehicle(s.ctx, 1, nil, nil, &last, &invalidCursor)
	s.Error(err)
	s.Contains(err.Error(), "illegal base64 data")
}

func (s *VehiclesSacdRepoTestSuite) TestCreateSacdResponse_InvalidPermissions() {
	_, grantee, err := helpers.GenerateWallet()
	s.NoError(err)

	currTime := time.Now().UTC().Truncate(time.Second)

	sacds := models.VehicleSacdSlice{
		&models.VehicleSacd{
			VehicleID:   1,
			Grantee:     grantee.Bytes(),
			Permissions: "invalid-binary", // Invalid binary string
			Source:      "test-source",
			CreatedAt:   currTime,
			ExpiresAt:   currTime.Add(time.Hour),
		},
	}

	pHelper := helpers.PaginationHelper[SacdCursor]{}

	result, err := s.repo.createSacdResponse(sacds, 1, false, false, pHelper)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "couldn't parse permission string")
}

func (s *VehiclesSacdRepoTestSuite) TestGetSacdsForVehicle_VehicleNotFound() {
	nonExistentVehicleID := 99999
	first := 10
	res, err := s.repo.GetSacdsForVehicle(s.ctx, nonExistentVehicleID, &first, nil, nil, nil)
	s.NoError(err)

	s.NotNil(res)
	s.Equal(0, res.TotalCount)
	s.Len(res.Edges, 0)
	s.Len(res.Nodes, 0)
}

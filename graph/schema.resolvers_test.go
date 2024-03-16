package graph

import (
	"context"
	"testing"

	"github.com/DIMO-Network/identity-api/graph/model"
	"github.com/DIMO-Network/identity-api/internal/repositories/aftermarket"
	"github.com/DIMO-Network/identity-api/internal/repositories/dcn"
	"github.com/DIMO-Network/identity-api/internal/repositories/manufacturer"
	"github.com/DIMO-Network/identity-api/internal/repositories/synthetic"
	"github.com/DIMO-Network/identity-api/internal/repositories/vehicle"
	"github.com/DIMO-Network/identity-api/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestQueryResolver_Node(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	testVehicle, err := vehicle.ToAPI(&models.Vehicle{ID: 1}, "", "")
	require.NoError(t, err)

	testAfterMarket, err := aftermarket.ToAPI(&models.AftermarketDevice{ID: 1}, "")
	require.NoError(t, err)

	testDCN, err := dcn.ToAPI(&models.DCN{Node: []byte{1, 2, 3, 4}})
	require.NoError(t, err)

	testManufacturer, err := manufacturer.ToAPI(&models.Manufacturer{ID: 1})
	require.NoError(t, err)

	testSynthetic, err := synthetic.ToAPI(&models.SyntheticDevice{ID: 1})
	require.NoError(t, err)

	// Define test cases
	testCases := []struct {
		name         string
		id           string
		setupMocks   func(*mockResolver)
		expectedNode model.Node
		hasError     bool
	}{
		{
			name: "vehicle",
			id:   testVehicle.ID,
			setupMocks: func(m *mockResolver) {
				m.mockVehicle.EXPECT().GetVehicle(ctx, testVehicle.TokenID).Return(testVehicle, nil)
			},
			expectedNode: testVehicle,
		},
		{
			name: "aftermarket",
			id:   testAfterMarket.ID,
			setupMocks: func(m *mockResolver) {
				by := model.AftermarketDeviceBy{TokenID: &testAfterMarket.TokenID}
				m.mockAftermarket.EXPECT().GetAftermarketDevice(ctx, by).Return(testAfterMarket, nil)
			},
			expectedNode: testAfterMarket,
		},
		{
			name: "dcn",
			id:   testDCN.ID,
			setupMocks: func(m *mockResolver) {
				by := model.DCNBy{Node: testDCN.Node}
				m.mockDCN.EXPECT().GetDCN(ctx, by).Return(testDCN, nil)
			},
			expectedNode: testDCN,
		},
		{
			name: "manufacturer",
			id:   testManufacturer.ID,
			setupMocks: func(m *mockResolver) {
				by := model.ManufacturerBy{TokenID: &testManufacturer.TokenID}
				m.mockManufacturer.EXPECT().GetManufacturer(ctx, by).Return(testManufacturer, nil)
			},
			expectedNode: testManufacturer,
		},
		{
			name: "synthetic",
			id:   testSynthetic.ID,
			setupMocks: func(m *mockResolver) {
				by := model.SyntheticDeviceBy{TokenID: &testSynthetic.TokenID}
				m.mockSynthetic.EXPECT().GetSyntheticDevice(ctx, by).Return(testSynthetic, nil)
			},
			expectedNode: testSynthetic,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Setup mocks
			mock := newMockResolver(ctrl)
			tc.setupMocks(mock)

			// Create the query resolver
			r := queryResolver{mock.Resolver()}

			// Run the test
			node, err := r.Node(ctx, tc.id)
			require.NoError(t, err)
			require.Equal(t, tc.expectedNode, node)
		})
	}
}

type mockResolver struct {
	mockVehicle      *MockVehicleRepository
	mockAftermarket  *MockAftermarketDeviceRepository
	mockDCN          *MockDCNRepository
	mockManufacturer *MockManufacturerRepository
	mockSynthetic    *MockSyntheticRepository
}

func newMockResolver(ctrl *gomock.Controller) *mockResolver {
	return &mockResolver{
		mockVehicle:      NewMockVehicleRepository(ctrl),
		mockAftermarket:  NewMockAftermarketDeviceRepository(ctrl),
		mockDCN:          NewMockDCNRepository(ctrl),
		mockManufacturer: NewMockManufacturerRepository(ctrl),
		mockSynthetic:    NewMockSyntheticRepository(ctrl),
	}
}

func (m *mockResolver) Resolver() *Resolver {
	return &Resolver{
		aftermarket:  m.mockAftermarket,
		dcn:          m.mockDCN,
		manufacturer: m.mockManufacturer,
		vehicle:      m.mockVehicle,
		synthetic:    m.mockSynthetic,
	}
}

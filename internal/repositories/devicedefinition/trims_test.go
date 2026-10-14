package devicedefinition

import (
	"testing"

	"github.com/DIMO-Network/identity-api/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ToAPI is the only place a trim reaches a client, so the mapping is worth
// pinning independently of the decode. No database or catalog is needed: the
// manufacturer argument is nil, which ToAPI already tolerates.
func TestToAPIMapsTrims(t *testing.T) {
	r := &Repository{}

	def := &services.CatalogDefinition{
		ID:    "toyota_camry_2020",
		Model: "Camry",
		Year:  2020,
		Metadata: services.CatalogDefinitionMetadata{DeviceAttributes: []services.CatalogDefinitionAttribute{
			{Name: "number_of_doors", Value: "4"},
		}},
		Trims: []services.CatalogTrim{
			{Name: "LE", Attributes: []services.CatalogDefinitionAttribute{
				{Name: "mpg_combined", Value: "34"},
				{Name: "powertrain_type", Value: "ICE"},
			}},
			{Name: "Hybrid LE", HardwareTemplateID: "42", Attributes: []services.CatalogDefinitionAttribute{
				{Name: "mpg_combined", Value: "52"},
			}},
		},
	}

	got, err := r.ToAPI(def, nil)
	require.NoError(t, err)

	// The flattened attributes are untouched -- this field is what every
	// existing client reads and it must not shift.
	require.Len(t, got.Attributes, 1)
	assert.Equal(t, "number_of_doors", got.Attributes[0].Name)

	require.Len(t, got.Trims, 2)

	assert.Equal(t, "LE", got.Trims[0].Name)
	assert.Nil(t, got.Trims[0].HardwareTemplateID, "absent rather than an empty string when the trim does not override it")
	require.Len(t, got.Trims[0].Attributes, 2)
	assert.Equal(t, "mpg_combined", got.Trims[0].Attributes[0].Name)
	assert.Equal(t, "34", got.Trims[0].Attributes[0].Value)

	assert.Equal(t, "Hybrid LE", got.Trims[1].Name)
	require.NotNil(t, got.Trims[1].HardwareTemplateID)
	assert.Equal(t, "42", *got.Trims[1].HardwareTemplateID)
}

// The dead-value filter the definition's own attributes get applies to trim
// attributes too: they come from the same documents.
func TestToAPIDropsDeadTrimAttributes(t *testing.T) {
	r := &Repository{}
	got, err := r.ToAPI(&services.CatalogDefinition{
		ID: "toyota_camry_2020", Model: "Camry", Year: 2020,
		Trims: []services.CatalogTrim{{Name: "LE", Attributes: []services.CatalogDefinitionAttribute{
			{Name: "mpg_combined", Value: "34"},
			{Name: "powertrain_type", Value: "<nil>"},
			{Name: "fuel_type", Value: ""},
			{Name: "", Value: "orphan"},
		}}},
	}, nil)
	require.NoError(t, err)

	require.Len(t, got.Trims, 1)
	require.Len(t, got.Trims[0].Attributes, 1)
	assert.Equal(t, "mpg_combined", got.Trims[0].Attributes[0].Name)
}

// A nameless trim is a malformed document, not an unnamed configuration: the
// worker's schema requires a name and it is how a client tells trims apart.
func TestToAPIDropsNamelessTrims(t *testing.T) {
	r := &Repository{}
	got, err := r.ToAPI(&services.CatalogDefinition{
		ID: "toyota_camry_2020", Model: "Camry", Year: 2020,
		Trims: []services.CatalogTrim{
			{Name: "", Attributes: []services.CatalogDefinitionAttribute{{Name: "mpg_combined", Value: "34"}}},
			{Name: "LE"},
		},
	}, nil)
	require.NoError(t, err)

	require.Len(t, got.Trims, 1)
	assert.Equal(t, "LE", got.Trims[0].Name)
}

// A legacy-manifest definition has no trims. The field is non-null in the
// schema, and gqlgen renders a nil slice as [], so this must stay nil-safe
// rather than being forced to an allocated empty slice somewhere.
func TestToAPIWithNoTrims(t *testing.T) {
	r := &Repository{}
	got, err := r.ToAPI(&services.CatalogDefinition{
		ID: "bmw-m_z4_2021", Model: "Z4", Year: 2021,
	}, nil)
	require.NoError(t, err)
	assert.Empty(t, got.Trims)
}

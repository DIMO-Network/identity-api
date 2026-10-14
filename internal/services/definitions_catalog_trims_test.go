package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The real shape definitions-worker serves, trimmed to what matters here. The
// template-level attributes are the three shared by every trim; the economy,
// drivetrain and powertrain values differ per trim and live on the trims. This
// is toyota_camry_2020 as the dev catalog actually holds it.
const camryTrimsTemplate = `{
  "id": "toyota_camry_2020",
  "deviceType": "vehicle",
  "manufacturer": {"slug": "toyota", "name": "Toyota", "tokenId": 128},
  "model": "Camry",
  "year": 2020,
  "attributes": {"emissions_standard": "Tier 3", "number_of_doors": 4, "vehicle_type": "PASSENGER CAR"},
  "trims": [
    {"name": "LE", "attributes": {"mpg_combined": 34, "wheelbase_in": 111.2, "powertrain_type": "ICE", "driven_wheels": "FWD"}},
    {"name": "Hybrid LE", "hardwareTemplateId": "42",
     "attributes": {"mpg_combined": 52, "wheelbase_in": 111.2, "powertrain_type": "HEV", "driven_wheels": "FWD"}}
  ],
  "version": 2,
  "createdAt": "2026-09-22T18:36:39.494Z",
  "updatedAt": "2026-09-22T19:04:17.586Z"
}`

func TestTemplateDecodeCarriesTrims(t *testing.T) {
	def, err := decodeTemplate(strings.NewReader(camryTrimsTemplate))
	require.NoError(t, err)

	require.Len(t, def.Trims, 2)

	le := def.Trims[0]
	assert.Equal(t, "LE", le.Name)
	assert.Empty(t, le.HardwareTemplateID, "this trim does not override the template's value")
	// Sorted by name, like the definition's own attributes, and rendered with
	// strconv so 111.2 does not become 1.112e+02.
	assert.Equal(t, []CatalogDefinitionAttribute{
		{Name: "driven_wheels", Value: "FWD"},
		{Name: "mpg_combined", Value: "34"},
		{Name: "powertrain_type", Value: "ICE"},
		{Name: "wheelbase_in", Value: "111.2"},
	}, le.Attributes)

	hybrid := def.Trims[1]
	assert.Equal(t, "Hybrid LE", hybrid.Name)
	assert.Equal(t, "42", hybrid.HardwareTemplateID)
	assert.Equal(t, "52", attrValue(t, hybrid.Attributes, "mpg_combined"))
	assert.Equal(t, "HEV", attrValue(t, hybrid.Attributes, "powertrain_type"))
}

// The point of the whole change: before trims were decoded, these values had no
// route to any consumer. The template is the only place they exist.
func TestTrimsCarryWhatFlatteningDrops(t *testing.T) {
	def, err := decodeTemplate(strings.NewReader(camryTrimsTemplate))
	require.NoError(t, err)

	flattened := map[string]bool{}
	for _, a := range def.Metadata.DeviceAttributes {
		flattened[a.Name] = true
	}
	for _, name := range []string{"mpg_combined", "wheelbase_in", "powertrain_type", "driven_wheels"} {
		assert.False(t, flattened[name], "%s varies by trim, so it must not be flattened onto the definition", name)
	}

	onTrims := map[string]bool{}
	for _, tr := range def.Trims {
		for _, a := range tr.Attributes {
			onTrims[a.Name] = true
		}
	}
	for _, name := range []string{"mpg_combined", "wheelbase_in", "powertrain_type", "driven_wheels"} {
		assert.True(t, onTrims[name], "%s must be reachable via trims", name)
	}
}

// Non-breaking: carrying trims through must not change the flattened attribute
// set by even one entry, because that is what every existing client reads.
func TestCarryingTrimsLeavesFlattenedAttributesAlone(t *testing.T) {
	def, err := decodeTemplate(strings.NewReader(camryTrimsTemplate))
	require.NoError(t, err)

	assert.Equal(t, []CatalogDefinitionAttribute{
		{Name: "emissions_standard", Value: "Tier 3"},
		{Name: "number_of_doors", Value: "4"},
		{Name: "vehicle_type", Value: "PASSENGER CAR"},
	}, def.Metadata.DeviceAttributes)
	assert.Equal(t, "toyota_camry_2020", def.ID)
	assert.Equal(t, "Camry", def.Model)
	assert.Equal(t, 2020, def.Year)
	assert.Equal(t, 128, def.Manufacturer.TokenID)
	assert.Empty(t, def.KSUID, "templates have no ksuid and one must not be invented")
}

// A single-trim template is still a trim, per the worker's schema, not a
// special case to collapse.
func TestSingleTrimTemplate(t *testing.T) {
	def, err := decodeTemplate(strings.NewReader(`{
      "id": "ford_f-150_2021", "deviceType": "vehicle",
      "manufacturer": {"slug": "ford", "name": "Ford", "tokenId": 40},
      "model": "F-150", "year": 2021,
      "attributes": {"number_of_doors": 4},
      "trims": [{"name": "Base", "attributes": {"mpg_combined": 22}}]
    }`))
	require.NoError(t, err)
	require.Len(t, def.Trims, 1)
	assert.Equal(t, "Base", def.Trims[0].Name)
}

// A template with no trims at all, and the legacy manifest which has no trims
// field, must both decode to an empty slice rather than tripping anything.
func TestTemplateWithoutTrims(t *testing.T) {
	def, err := decodeTemplate(strings.NewReader(`{
      "id": "toyota_camry_2020", "deviceType": "vehicle",
      "manufacturer": {"slug": "toyota", "name": "Toyota", "tokenId": 128},
      "model": "Camry", "year": 2020, "attributes": {"number_of_doors": 4}
    }`))
	require.NoError(t, err)
	assert.NotNil(t, def.Trims, "an absent trims array decodes to empty, never nil")
	assert.Empty(t, def.Trims)
}

func attrValue(t *testing.T, attrs []CatalogDefinitionAttribute, name string) string {
	t.Helper()
	for _, a := range attrs {
		if a.Name == name {
			return a.Value
		}
	}
	t.Fatalf("attribute %q not found", name)
	return ""
}

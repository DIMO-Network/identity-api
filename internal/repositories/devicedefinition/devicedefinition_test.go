package devicedefinition

import (
	"testing"
)

const migrationsDir = "../../../migrations"

func TestGetDeviceDefinition(t *testing.T) {
	//ctx := context.Background()
	//
	//pdb, _ := helpers.StartContainerDatabase(ctx, t, migrationsDir)
	//manufacturers := []string{"ford", "tesla", "kia", "acura", "honda", "jeep"}
	//
	//for i := 0; i < 6; i++ {
	//	m := models.Manufacturer{
	//		ID:       i,
	//		Name:     manufacturers[i],
	//		Owner:    common.FromHex("3232323232323232323232323232323232323232"),
	//		MintedAt: time.Now(),
	//	}
	//
	//	err := m.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	//	assert.NoError(t, err)
	//}
	//
	//logger := zerolog.Nop()
	//baseRepo := base.NewRepository(pdb, config.Settings{}, &logger)
	//
	//controller := Repository{}
	//for i := 0; i < 6; i++ {
	//	slug := helpers.SlugString(manufacturers[i])
	//
	//	res, err := controller.GetDeviceDefinition(ctx, model.DevicedefinitionBy{ID: slug})
	//	assert.NoError(t, err)
	//	assert.Equal(t, res.ID, slug)
	//}

}

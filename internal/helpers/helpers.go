package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
)

func CursorToID(cur string) (int, error) {
	b, err := base64.StdEncoding.DecodeString(cur)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(b))
}

func IDToCursor(id int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(id)))
}

func WithSchema(tableName string) string {
	return "identity_api." + tableName
}

func GenerateDCNNode() []byte {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

func GetVehicleImageUrl(baseURL string, tokenID int) string {
	return fmt.Sprintf("%s/vehicle/%d/image", baseURL, tokenID)
}

func GetAftermarketDeviceImageUrl(baseURL string, tokenID int) string {
	return fmt.Sprintf("%s/aftermarket/device/%d/image", baseURL, tokenID)
}

func GetVehicleDataURI(baseURL string, tokenID int) string {
	return fmt.Sprintf("%s%d", baseURL, tokenID)
}

package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

func SlugString(term string) string {

	lowerCase := cases.Lower(language.English, cases.NoLower)
	lowerTerm := lowerCase.String(term)

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	cleaned, _, _ := transform.String(t, lowerTerm)
	cleaned = strings.ReplaceAll(cleaned, " ", "-")
	cleaned = strings.ReplaceAll(cleaned, "_", "-")

	return cleaned

}

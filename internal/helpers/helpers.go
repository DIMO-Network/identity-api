package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/DIMO-Network/go-mnemonic"
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

func CreateMnemonic(addr []byte) (string, error) {
	mn, err := mnemonic.EntropyToMnemonicThreeWords(addr)
	if err != nil {
		return "", err
	}
	name := strings.Join(mn, " ")

	return name, nil
}

func IntToBytes(intVal int) []byte {
	b := make([]byte, 32)
	binary.LittleEndian.PutUint32(b, uint32(intVal))

	return b
}

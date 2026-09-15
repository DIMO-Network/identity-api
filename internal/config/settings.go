package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/DIMO-Network/shared/pkg/db"
)

// Settings contains the application config
type Settings struct {
	LogLevel              string      `yaml:"LOG_LEVEL"`
	DB                    db.Settings `yaml:"DB"`
	Port                  int         `yaml:"PORT"`
	MonPort               int         `yaml:"MON_PORT"`
	KafkaBrokers          string      `yaml:"KAFKA_BROKERS"`
	ContractsEventTopic   string      `yaml:"CONTRACT_EVENT_TOPIC"`
	DIMORegistryChainID   int64       `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	DIMORegistryAddr      string      `yaml:"DIMO_REGISTRY_ADDR"`
	VehicleNFTAddr        string      `yaml:"DIMO_VEHICLE_NFT_ADDR"`
	ManufacturerNFTAddr   string      `yaml:"DIMO_MANUFACTURER_NFT_ADDR"`
	AftermarketDeviceAddr string      `yaml:"AFTERMARKET_DEVICE_CONTRACT_ADDRESS"`
	SACDAddress           string      `yaml:"SACD_ADDRESS"`
	DCNRegistryAddr       string      `yaml:"DCN_REGISTRY_ADDR"`
	DCNResolverAddr       string      `yaml:"DCN_RESOLVER_ADDR"`
	SyntheticDeviceAddr   string      `yaml:"SYNTHETIC_DEVICE_CONTRACT_ADDRESS"`
	RewardsContractAddr   string      `yaml:"REWARDS_CONTRACT_ADDRESS"`
	BaseImageURL          string      `yaml:"BASE_IMAGE_URL"`
	BaseVehicleDataURI    string      `yaml:"BASE_VEHICLE_DATA_URI"`
	DefinitionsCatalogURL string      `yaml:"DEFINITIONS_CATALOG_URL"`
	// Smallest valid-definition count worth believing. A catalog carrying
	// fewer is refused on every pod, cold or warm. Empty or "0" disables the
	// check. Held as a string because the shared loader turns an unparseable
	// int into 0 without an error, which silently disabled the floor;
	// DefinitionsCatalog parses it strictly.
	DefinitionsMinCount string `yaml:"DEFINITIONS_MIN_COUNT"`
	// How old the catalog a replica serves may get while refreshes fail. Past
	// this, device-definition queries fail instead of quietly serving data
	// that stopped changing. A Go duration string; empty means 24h and "0"
	// disables the bound.
	DefinitionsMaxStaleness string `yaml:"DEFINITIONS_MAX_STALENESS"`
	EthereumRPCURL          string `yaml:"ETHEREUM_RPC_URL"`
	DevLicenseAddr          string `yaml:"DEV_LICENSE_ADDR"`
	StakingAddr             string `yaml:"STAKING_ADDR"`
	ConnectionAddr          string `yaml:"CONNECTION_ADDR"`
	StorageNodeAddr         string `yaml:"STORAGE_NODE_ADDR"`
	TemplateAddr            string `yaml:"TEMPLATE_ADDR"`
	FetchAPIGRPCAddr        string `yaml:"FETCH_API_GRPC_ADDR"`
}

// DefaultDefinitionsMaxStaleness applies when DEFINITIONS_MAX_STALENESS is unset.
const DefaultDefinitionsMaxStaleness = 24 * time.Hour

// DefinitionsCatalogConfig is the validated form of the DEFINITIONS_* settings.
type DefinitionsCatalogConfig struct {
	// URL is the catalog base URL with any trailing slashes removed.
	URL string
	// MinCount is the valid-definition floor. Zero disables it.
	MinCount int
	// MaxStaleness bounds how old a served catalog may be. Zero disables it.
	MaxStaleness time.Duration
}

// DefinitionsCatalog validates the DEFINITIONS_* settings and returns their
// parsed form. The shared loader cannot: it assigns strings verbatim, so a
// quoted Helm value with a trailing space or newline reaches the pod intact,
// and it turns a malformed int into 0 with no error. Call it at startup and
// refuse to start on an error.
func (s *Settings) DefinitionsCatalog() (DefinitionsCatalogConfig, error) {
	catalogURL, err := parseDefinitionsCatalogURL(s.DefinitionsCatalogURL)
	if err != nil {
		return DefinitionsCatalogConfig{}, err
	}
	minCount, err := parseDefinitionsMinCount(s.DefinitionsMinCount)
	if err != nil {
		return DefinitionsCatalogConfig{}, err
	}
	maxStaleness, err := parseDefinitionsMaxStaleness(s.DefinitionsMaxStaleness)
	if err != nil {
		return DefinitionsCatalogConfig{}, err
	}
	return DefinitionsCatalogConfig{URL: catalogURL, MinCount: minCount, MaxStaleness: maxStaleness}, nil
}

func parseDefinitionsCatalogURL(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("DEFINITIONS_CATALOG_URL is required")
	}
	if i := strings.IndexFunc(raw, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }); i >= 0 {
		return "", fmt.Errorf("DEFINITIONS_CATALOG_URL %q contains whitespace or a control character at byte %d", raw, i)
	}
	// Every neighbouring catalog setting in values.yaml carries a trailing
	// slash, and "//manifest.json" does not match the worker's exact routes.
	trimmed := strings.TrimRight(raw, "/")
	// Paths are appended to the base, so a query or fragment would swallow them.
	if strings.ContainsAny(trimmed, "?#") {
		return "", fmt.Errorf("DEFINITIONS_CATALOG_URL %q must not carry a query or fragment", raw)
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("DEFINITIONS_CATALOG_URL %q is not a valid URL: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("DEFINITIONS_CATALOG_URL %q must use http or https", raw)
	}
	if u.Host == "" {
		return "", fmt.Errorf("DEFINITIONS_CATALOG_URL %q has no host", raw)
	}
	return trimmed, nil
}

func parseDefinitionsMinCount(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("DEFINITIONS_MIN_COUNT %q must be a non-negative base-10 integer: digits only, with no sign, separator or whitespace", raw)
		}
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("DEFINITIONS_MIN_COUNT %q: %w", raw, err)
	}
	return n, nil
}

func parseDefinitionsMaxStaleness(raw string) (time.Duration, error) {
	if raw == "" {
		return DefaultDefinitionsMaxStaleness, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("DEFINITIONS_MAX_STALENESS %q must be a Go duration such as 24h, or 0 to disable the bound: %w", raw, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("DEFINITIONS_MAX_STALENESS %q must not be negative", raw)
	}
	return d, nil
}

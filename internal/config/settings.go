package config

import "github.com/DIMO-Network/shared/pkg/db"

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
	TablelandAPIGateway   string      `yaml:"TABLELAND_API_GATEWAY"`
	EthereumRPCURL        string      `yaml:"ETHEREUM_RPC_URL"`
	DevLicenseAddr        string      `yaml:"DEV_LICENSE_ADDR"`
	StakingAddr           string      `yaml:"STAKING_ADDR"`
	ConnectionAddr        string      `yaml:"CONNECTION_ADDR"`
	StorageNodeAddr       string      `yaml:"STORAGE_NODE_ADDR"`
	TemplateAddr          string      `yaml:"TEMPLATE_ADDR"`
}

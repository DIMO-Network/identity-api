package config

import "github.com/DIMO-Network/shared/db"

// Settings contains the application config
type Settings struct {
	DB                    db.Settings `yaml:"DB"`
	Port                  int         `yaml:"PORT"`
	KafkaBrokers          string      `yaml:"KAFKA_BROKERS"`
	ContractsEventTopic   string      `yaml:"CONTRACT_EVENT_TOPIC"`
	DIMORegistryChainID   int64       `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	DIMORegistryAddr      string      `yaml:"DIMO_REGISTRY_ADDR"`
	VehicleNFTAddr        string      `yaml:"DIMO_VEHICLE_NFT_ADDR"`
	AftermarketDeviceAddr string      `yaml:"AFTERMARKET_DEVICE_CONTRACT_ADDRESS"`
	DCNRegistryAddr       string      `yaml:"DCN_REGISTRY_ADDR"`
	DCNResolverAddr       string      `yaml:"DCN_RESOLVER_ADDR"`
}

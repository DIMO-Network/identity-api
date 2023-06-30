package config

import "github.com/DIMO-Network/shared/db"

// Settings contains the application config
type Settings struct {
	DB                  db.Settings `yaml:"DB"`
	Port                int         `yaml:"PORT"`
	KafkaBrokers        string      `yaml:"KAFKA_BROKERS"`
	ContractsEventTopic string      `yaml:"CONTRACT_EVENT_TOPIC"`
	DIMORegistryChainID int64       `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	DefinitionsGRPCAddr string      `yaml:"DEFINITIONS_GRPC_ADDR"`
	DevicesApiGrpcAddr  string      `yaml:"DEVICES_API_GRPC_ADDR"`
}

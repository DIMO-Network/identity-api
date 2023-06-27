package config

import "github.com/DIMO-Network/shared/db"

// Settings contains the application config
type Settings struct {
	DB   db.Settings `yaml:"DB"`
	Port int         `json:"PORT"`
}

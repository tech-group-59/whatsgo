package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type GoogleCloudConfig struct {
	Enabled         bool   `yaml:"enabled"`
	CredentialsFile string `yaml:"credentials_file"`
	FolderID        string `yaml:"folder_id"`
}

type Config struct {
	Chats              []string          `yaml:"chats"`
	FileStoragePath    string            `yaml:"file_storage_path"`
	DBConnectionString string            `yaml:"db_connection_string"`
	DBDialect          string            `yaml:"db_dialect"`
	Debug              bool              `yaml:"debug"`
	GoogleCloud        GoogleCloudConfig `yaml:"google_cloud"`
}

func LoadConfig(file string) (*Config, error) {
	var config Config
	configFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) IsChatTrackable(chatID string) bool {
	if c.Chats == nil {
		return true
	}

	for _, id := range c.Chats {
		if id == chatID {
			return true
		}
	}
	return false
}

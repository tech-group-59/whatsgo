package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type CSVConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type DBConfig struct {
	ConnectionString string `yaml:"connection_string"`
	Dialect          string `yaml:"dialect"`
}

type GoogleCloudConfig struct {
	Enabled         bool   `yaml:"enabled"`
	CredentialsFile string `yaml:"credentials_file"`
	TokenFile       string `yaml:"token_file"`
	FolderID        string `yaml:"folder_id"`
}

type OCRConfig struct {
	Enabled bool `yaml:"enabled"`
}

type WebhookConfig struct {
	Enabled bool     `yaml:"enabled"`
	URLs    []string `yaml:"urls"`
}

type Chat struct {
	ID    string `yaml:"id"`
	Alias string `yaml:"alias,omitempty"`
}

type Config struct {
	Chats           []Chat            `yaml:"chats"`
	FileStoragePath string            `yaml:"file_storage_path"`
	Database        DBConfig          `yaml:"database"`
	CSV             CSVConfig         `yaml:"csv"`
	GoogleCloud     GoogleCloudConfig `yaml:"google_cloud"`
	OCR             OCRConfig         `yaml:"ocr"`
	Webhook         WebhookConfig     `yaml:"webhook"`
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
	log.Infof("Trackable chats: %v", config.Chats)
	return &config, nil
}

func (c *Config) IsChatTrackable(chatID string) bool {
	if c.Chats == nil {
		return true
	}

	log.Infof("Checking if chat '%s' is trackable", chatID)
	for _, chat := range c.Chats {
		if chat.ID == chatID {
			return true
		}
	}
	return false
}

func GetDefaultConfig() *Config {
	return &Config{
		Chats:           nil,
		FileStoragePath: "file-storage",
		Database: DBConfig{
			Dialect:          "sqlite3",
			ConnectionString: "file:whatsgo.db?_foreign_keys=on",
		},
		CSV: CSVConfig{
			Enabled: false,
			Path:    "csv",
		},
		GoogleCloud: GoogleCloudConfig{
			Enabled:         false,
			CredentialsFile: "credentials.json",
			TokenFile:       "token.json",
		},
		OCR: OCRConfig{
			Enabled: false,
		},
	}
}

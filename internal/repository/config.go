package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)


type Config struct {
	ClientID     string `json:"client_id"`
}

type ConfigRepository struct {
	path string
}

func NewConfigRepository(path string) *ConfigRepository {
	return &ConfigRepository{path: path}
}

func (r *ConfigRepository) Load() (*Config, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ConfigRepository) Save(cfg *Config) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0600)
}

func (r *ConfigRepository) Delete() error {
	return os.Remove(r.path)
}

func (r *ConfigRepository) Exists() bool {
	_, err := os.Stat(r.path)
	return err == nil
}
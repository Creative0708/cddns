package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/adrg/xdg"
)

type Config struct {
	ServerLocation string
	ApiKey         string
}

func configPath() (string, error) {
	return xdg.ConfigFile("cddns/config.json")
}

func loadConfig() (*Config, error) {
	filePath, err := configPath()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}

	err = validateApiKey(cfg.ApiKey)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func saveConfig(config *Config) error {
	filePath, err := configPath()
	if err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, bytes, 0o600)

	if err != nil {
		log.Print("warning: failed to save config: ", err)
	} else {
		fmt.Print("saved config to ", filePath)
	}
	return nil
}

func getPermanentBinaryPath() string {
	os.MkdirAll(xdg.BinHome, 0o755)
	permanentPath := path.Join(xdg.BinHome, "cddns")

	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("failed to get path to executable: ", err)
	}
	contents, err := os.ReadFile(exePath)
	if err != nil {
		log.Fatal("failed to get read executable: ", err)
	}

	os.WriteFile(permanentPath, contents, 0o755)
	return permanentPath
}

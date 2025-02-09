package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	var conf = Config{}
	configPath, err := getConfigPath()
	if err != nil {
		return conf, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return conf, err
	}

	if err = json.Unmarshal(data, &conf); err != nil {
		return conf, err
	}
	return conf, nil
}

func (cnf *Config) SetUser(name string) error {
	cnf.CurrentUserName = name
	writeConfig(cnf)
	return nil
}

func writeConfig(cfg *Config) error {
	byteData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}
	err = os.WriteFile(cfgPath, byteData, 0666)
	return err
}

func getConfigPath() (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := homeDir + "/.gatorconfig.json"
	return configPath, nil
}

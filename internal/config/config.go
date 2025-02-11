package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
	ConnectString   string `json:"connect_string"`
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

func (conf *Config) SetUser(name string) error {
	conf.CurrentUserName = name
	writeConfig(conf)
	return nil
}

func writeConfig(conf *Config) error {
	byteData, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	confPath, err := getConfigPath()
	if err != nil {
		return err
	}
	err = os.WriteFile(confPath, byteData, 0666)
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

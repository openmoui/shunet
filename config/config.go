package config

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"shunet/utils"
	"sync"
)

var log = utils.Log

type Config struct {
	UserId            string `yaml:"userId"` // 学号
	Password          string `yaml:"password"`
	PublicKeyExponent string `yaml:"publicKeyExponent,omitempty"`
	PublicKeyModulus  string `yaml:"publicKeyModulus,omitempty"`
	PasswordEncrypt   string `yaml:"-"`
	Mac               string `yaml:"mac,omitempty"`
	Host              string `yaml:"host,omitempty"`
	DelayTime         int    `yaml:"delayTime,omitempty"`
	Pid               int    `yaml:"pid,omitempty"`
	LogLevel          string `yaml:"logLevel,omitempty"`
	Proxy             string `yaml:"proxy,omitempty"`
	filePath          string `yaml:"-"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	config.filePath = path
	utils.SetLogLevel(config.LogLevel)

	// 设置默认值
	if len(config.Host) == 0 {
		config.Host = "10.10.9.9"
	}
	config.PasswordEncrypt = "true"
	return &config, nil
}

func (c *Config) Save() error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	file, err := os.Create(c.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(bytes); err != nil {
		log.Errorf("config save Write err: %+v", err)
		return err
	}

	if err = file.Sync(); err != nil {
		log.Warning("config save Sync err: %+v", err)
	}
	log.Info("Config.Save Save config")
	return nil
}

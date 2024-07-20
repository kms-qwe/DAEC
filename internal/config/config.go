package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string        `yaml:"env" env-default:"local"`
	StoragePath    string        `yaml:"storage_path" env-required:"true"`
	TokenTTL       time.Duration `yaml:"tokenTTL" env-required:"true"`
	GRPC           GRPCConfig    `yaml:"grpc_server" env-required:"true"`
	HTTP           HTTPConfig    `yaml:"http_server" env-required:"true"`
	Addition       time.Duration `yaml:"time_addition_ms" env-required:"true"`
	Subtraction    time.Duration `yaml:"time_subtraction_ms" env-required:"true"`
	Multiplication time.Duration `yaml:"time_multiplication_ms" env-required:"true"`
	Division       time.Duration `yaml:"time_division_ms" env-required:"true"`
	ComputingPower int           `yaml:"computing_power" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type HTTPConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MastLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

package application

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort   uint16
	RedisAddress string
}

func LoadConfig() Config {
	cfg := Config{
		ServerPort:   3000,
		RedisAddress: "localhost:6379",
	}

	if ServerPort, exits := os.LookupEnv("SERVER_PORT"); exits {
		if value, err := strconv.ParseUint(ServerPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(value)
		}
	}

	if RedisPort, exits := os.LookupEnv("REDIS_PORT"); exits {
		cfg.RedisAddress = RedisPort
	}

	return cfg
}

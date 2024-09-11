package config

import (
	"github.com/spf13/viper"
	"time"
)

type (
	TokenRps struct {
		Values map[string]int `mapstructure:",remain"`
	}
	Config struct {
		Port      int           `mapstructure:"PORT"`
		RedisHost string        `mapstructure:"REDIS_HOST"`
		RedisPort int           `mapstructure:"REDIS_PORT"`
		Rps       int           `mapstructure:"RPS"`
		Interval  time.Duration `mapstructure:"INTERVAL"`
		TokenRps  TokenRps      `mapstructure:"TOKEN_RPS"`
	}
)

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")

	viper.SetDefault("PORT", 8080)
	viper.SetDefault("REDIS_HOST", "127.0.0.1")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("RPS", 10)
	viper.SetDefault("INTERVAL", time.Minute*5)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

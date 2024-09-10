package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setup(t *testing.T, keyValueItems ...string) {
	fs := afero.NewMemMapFs()
	viper.SetFs(fs)

	f, err := fs.Create(".env")
	if assert.NoError(t, err) {
		defer f.Close()
		for _, keyValue := range keyValueItems {
			_, _ = f.Write([]byte(keyValue + "\n"))
		}
	}
}

func TestLoadConfig(t *testing.T) {
	setup(t,
		"REDIS_HOST=192.168.1.1",
		"REDIS_PORT=6666",
		"RPS=20",
		"INTERVAL=35s",
		"TOKEN_RPS.abc123=20",
	)

	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.1", cfg.RedisHost)
	assert.Equal(t, 6666, cfg.RedisPort)
	assert.Equal(t, 20, cfg.Rps)
	assert.Equal(t, time.Second*35, cfg.Interval)
	assert.Equal(t, map[string]int{"abc123": 20}, cfg.TokenRps.Values)
}

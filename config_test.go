package bramble

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("port provided", func(t *testing.T) {
		cfg := &Config{
			GatewayPort: defaultPortGateway,
			PrivatePort: 8083,
			MetricsPort: 8084,
		}
		require.Equal(t, fmt.Sprintf(":%d", defaultPortGateway), cfg.GatewayAddress())
		require.Equal(t, ":8083", cfg.PrivateAddress())
		require.Equal(t, ":8084", cfg.MetricAddress())
	})
	t.Run("address provided and prefered over port", func(t *testing.T) {
		cfg := &Config{
			GatewayListenAddress: defaultAddressGateway,
			GatewayPort:          0,
			PrivateListenAddress: "127.0.0.1:" + strconv.Itoa(8084),
			PrivatePort:          defaultPortPrivate,
			MetricsListenAddress: "",
			MetricsPort:          8084,
		}
		require.Equal(t, "0.0.0.0:8082", cfg.GatewayAddress())
		require.Equal(t, "127.0.0.1:8084", cfg.PrivateAddress())
		require.Equal(t, ":8084", cfg.MetricAddress())
	})
	t.Run("private http address for plugin services", func(t *testing.T) {
		cfg := newConfig()
		require.Equal(t, fmt.Sprintf("http://localhost:%d/plugin", defaultPortPrivate), cfg.PrivateHttpAddress("plugin"))
		cfg.PrivateListenAddress = "127.0.0.1:8084"
		require.Equal(t, "http://127.0.0.1:8084/plugin", cfg.PrivateHttpAddress("plugin"))
	})
	t.Run("default loglevel", func(t *testing.T) {
		cfg := newConfig()
		require.Equal(t, log.DebugLevel, cfg.LogLevel)
	})
	t.Run("loglevel set from environment", func(t *testing.T) {
		require.NoError(t, os.Setenv(envBrambleLogLevel, "ERROR"))
		cfg := newConfig()
		cfg.Services = []string{"127.0.0.1:9999"}
		require.NoError(t, cfg.Load())
		require.Equal(t, log.ErrorLevel, cfg.LogLevel)
	})
}

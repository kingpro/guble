package config

import (
	"os"
	"testing"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/stretchr/testify/assert"
)

func TestParsingOfEnviromentVariables(t *testing.T) {
	a := assert.New(t)

	originalArgs := os.Args
	os.Args = []string{os.Args[0]}
	defer func() { os.Args = originalArgs }()

	// given: some environment variables
	os.Setenv("GUBLE_LISTEN", "listen")
	defer os.Unsetenv("GUBLE_LISTEN")

	os.Setenv("GUBLE_LOG_INFO", "true")
	defer os.Unsetenv("GUBLE_LOG_INFO")

	os.Setenv("GUBLE_LOG_DEBUG", "true")
	defer os.Unsetenv("GUBLE_LOG_DEBUG")

	os.Setenv("GUBLE_KV_BACKEND", "kv-backend")
	defer os.Unsetenv("GUBLE_KV_BACKEND")

	os.Setenv("GUBLE_STORAGE_PATH", "storage-path")
	defer os.Unsetenv("GUBLE_STORAGE_PATH")

	os.Setenv("GUBLE_HEALTH_ENDPOINT", "health_endpoint")
	defer os.Unsetenv("GUBLE_HEALTH_ENDPOINT")

	os.Setenv("GUBLE_METRICS_ENDPOINT", "metrics_endpoint")
	defer os.Unsetenv("GUBLE_METRICS_ENDPOINT")

	os.Setenv("GUBLE_MS_BACKEND", "ms-backend")
	defer os.Unsetenv("GUBLE_MS_BACKEND")

	os.Setenv("GUBLE_GCM_ENABLED", "true")
	defer os.Unsetenv("GUBLE_GCM_ENABLED")

	os.Setenv("GUBLE_GCM_API_KEY", "gcm-api-key")
	defer os.Unsetenv("GUBLE_GCM_API_KEY")

	os.Setenv("GUBLE_GCM_WORKERS", "3")
	defer os.Unsetenv("GUBLE_GCM_WORKERS")

	// when we parse the arguments
	kingpin.Parse()

	// the the arg parameters are set
	assertArguments(a)
}

func TestParsingArgs(t *testing.T) {
	a := assert.New(t)

	originalArgs := os.Args

	defer func() { os.Args = originalArgs }()

	// given: a command line
	os.Args = []string{os.Args[0],
		"--listen", "listen",
		"--log-info",
		"--log-debug",
		"--kv-backend", "kv-backend",
		"--storage-path", "storage-path",
		"--ms-backend", "ms-backend",
		"--health", "health_endpoint",
		"--metrics", "metrics_endpoint",
		"--gcm-enabled",
		"--gcm-api-key", "gcm-api-key",
		"--gcm-workers", "3",
	}

	// when we parse the arguments
	kingpin.Parse()

	// the the arg parameters are set
	assertArguments(a)
}

func assertArguments(a *assert.Assertions) {
	a.Equal("listen", *Listen)
	a.Equal("kv-backend", *KVBackend)
	a.Equal("storage-path", *StoragePath)
	a.Equal("ms-backend", *MSBackend)

	a.Equal("health_endpoint", *Health)
	a.Equal("metrics_endpoint", *Metrics)

	a.Equal(true, *GCM.Enabled)
	a.Equal("gcm-api-key", *GCM.APIKey)
	a.Equal(3, *GCM.Workers)

	a.Equal(true, *Log.Info)
	a.Equal(true, *Log.Debug)

}

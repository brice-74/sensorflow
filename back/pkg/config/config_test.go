package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/brice-74/sensorflow/pkg/config"
	"github.com/stretchr/testify/require"
)

func resetEnvAndFlags() {
	os.Clearenv()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestLoader(t *testing.T) {
	t.Run("EnvOnly_String", func(t *testing.T) {
		resetEnvAndFlags()
		os.Setenv("MY_KEY", "hello")

		loader := config.NewLoader(config.EnvOnly)
		var value string
		loader.String(&value, "key", "MY_KEY", "default", "test key")

		err := loader.Parse()
		require.NoError(t, err)
		require.Equal(t, "hello", value)
	})

	t.Run("FlagOnly_Int", func(t *testing.T) {
		resetEnvAndFlags()
		os.Args = []string{"cmd", "-port=1234"}

		loader := config.NewLoader(config.FlagOnly)
		var port int
		loader.Int(&port, "port", "PORT", 80, "server port")

		err := loader.Parse()
		require.NoError(t, err)
		require.Equal(t, 1234, port)
	})

	t.Run("Both_OverrideByFlag", func(t *testing.T) {
		resetEnvAndFlags()
		os.Setenv("PORT", "9999")
		os.Args = []string{"cmd", "-port=5678"}

		loader := config.NewLoader(config.Both)
		var port int
		loader.Int(&port, "port", "PORT", 80, "server port")

		err := loader.Parse()
		require.NoError(t, err)
		require.Equal(t, 5678, port)
	})

	t.Run("DefaultValue_String", func(t *testing.T) {
		resetEnvAndFlags()

		loader := config.NewLoader(config.EnvOnly)
		var name string
		loader.String(&name, "name", "NAME", "defaultName", "user name")

		err := loader.Parse()
		require.NoError(t, err)
		require.Equal(t, "defaultName", name)
	})

	t.Run("RequiredMissing", func(t *testing.T) {
		resetEnvAndFlags()

		loader := config.NewLoader(config.EnvOnly)
		var name string
		loader.String(&name, "name", "NAME", "", "required").Required()

		err := loader.Parse()
		require.Error(t, err)
		require.Zero(t, name)
	})

	t.Run("InvalidEnvParsing", func(t *testing.T) {
		resetEnvAndFlags()
		os.Setenv("PORT", "notanumber")

		loader := config.NewLoader(config.EnvOnly)
		var port int
		loader.Int(&port, "port", "PORT", 42, "test")

		err := loader.Parse()
		require.Error(t, err)
		require.Zero(t, port)
	})
}

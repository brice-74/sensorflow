//go:build unit

package zerolog_test

import (
	"bytes"
	"errors"
	"testing"

	zerologadapter "github.com/brice-74/sensorflow/internal/adapters/zerolog"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestZeroLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := zerologadapter.NewLogger(zerolog.New(&buf))

	t.Run("Info", func(t *testing.T) {
		logger.Info("test message")

		output := buf.String()
		require.Contains(t, output, `"level":"info"`)
		require.Contains(t, output, `"message":"test message"`)
		buf.Reset()
	})

	t.Run("Error", func(t *testing.T) {
		logger.Error(errors.New("test error"))

		output := buf.String()
		require.Contains(t, output, `"level":"error"`)
		require.Contains(t, output, `"error":"test error"`)
		buf.Reset()
	})

	t.Run("Fatal", func(t *testing.T) {
		logger.Fatal(errors.New("fatal error"))

		output := buf.String()
		require.Contains(t, output, `"level":"fatal"`)
		require.Contains(t, output, `"error":"fatal error"`)
		buf.Reset()
	})

	t.Run("With", func(t *testing.T) {
		subLogger := logger.With(
			&log.User{
				ID:       "u123",
				Email:    "user@example.com",
				Username: "testuser",
			},
			log.Tags{
				"env": "production",
			},
			log.Contexts{
				"request": {"id": "123"},
			},
		)
		subLogger.Info("message with options")

		output := buf.String()
		require.Contains(t, output, `"ctx_request":{"id":"123"}`)
		require.Contains(t, output, `"tag_env":"production"`)
		require.Contains(t, output, `"user":{"id":"u123","email":"user@example.com","username":"testuser","data":null}`)
		require.Contains(t, output, `"message":"message with options"`)
		buf.Reset()
	})

	t.Run("ChildDoNotAffectParent", func(t *testing.T) {
		_ = logger.With(
			log.Contexts{
				"request": {"id": "123"},
			},
		)
		logger.Info("parent message")

		output := buf.String()
		require.NotContains(t, output, `"ctx_request":{"id":"123"}`)
		buf.Reset()
	})
}

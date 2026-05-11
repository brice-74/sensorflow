package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidationHelpers(t *testing.T) {
	t.Run("IsGreaterEqualThan", func(t *testing.T) {
		validate := IsGreaterEqualThan(10)
		require.NoError(t, validate(10))
		require.NoError(t, validate(11))
		require.Error(t, validate(9))
	})

	t.Run("IsGreaterThan", func(t *testing.T) {
		validate := IsGreaterThan(10)
		require.NoError(t, validate(11))
		require.Error(t, validate(10))
	})

	t.Run("IsLessThan", func(t *testing.T) {
		validate := IsLessThan(10)
		require.NoError(t, validate(9))
		require.Error(t, validate(10))
	})

	t.Run("IsLessEqualThan", func(t *testing.T) {
		validate := IsLessEqualThan(10)
		require.NoError(t, validate(10))
		require.Error(t, validate(11))
	})

	t.Run("IsBetween", func(t *testing.T) {
		validate := IsBetween(1, 5)
		require.NoError(t, validate(1))
		require.NoError(t, validate(3))
		require.NoError(t, validate(5))
		require.Error(t, validate(0))
		require.Error(t, validate(6))
	})

	t.Run("IsNonEmptyString", func(t *testing.T) {
		require.NoError(t, IsNonEmptyString("hello"))
		require.Error(t, IsNonEmptyString(""))
	})

	t.Run("IsPortString", func(t *testing.T) {
		require.NoError(t, IsPortString("8080"))
		require.Error(t, IsPortString(""))
		require.Error(t, IsPortString("abc"))
		require.Error(t, IsPortString("70000"))
	})
}

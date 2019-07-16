package lash_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
)

func Test_env_vars(t *testing.T) {
	t.Run("missing vars are treated as errors", func(t *testing.T) {
		session := lash.NewSession()
		session.OnError(lash.Ignore)
		session.Env().Require("non-existing", "helpful message")

		assert.Error(t, session.Err())
		assert.Contains(t, session.ErrorString(), "Env:Require:missing")
		assert.Contains(t, session.ErrorString(), "non-existing")
		assert.Contains(t, session.ErrorString(), "helpful message")
	})
	t.Run("existing vars are not treated as errors", func(t *testing.T) {
		session := lash.NewSession()
		session.OnError(lash.Ignore)
		require.NoError(t, os.Setenv("existing-key", "a-value"))

		session.Env().Require("existing-key", "helpful message")

		assert.NoError(t, session.Err())
	})
}

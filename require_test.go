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
		require.NoError(t, os.Setenv("another-key", "another-value"))

		session.
			Env().
			Require("existing-key", "helpful message").
			Require("another-key", "helpful message")

		assert.NoError(t, session.Err())
	})
}

func Test_flags(t *testing.T) {
	t.Run("flags can be required", func(t *testing.T) {
		session := lash.NewSession()
		session.OnError(lash.Ignore)

		args := session.Args()

		// Args are automatically set  by calling Args above
		require.NotEmpty(t, args)
		// overwrite for the test
		args.Args = []string{"the-program", "first", "second"}

		args.
			Require(1, "helpful text 1").
			Require(2, "helpful text 2")
		assert.NoError(t, session.Err())

		args.Require(3, "helpful text 3")
		assert.Error(t, session.Err())

		assert.Contains(t, session.ErrorString(), "Arg:Require:missing")
		assert.Contains(t, session.ErrorString(), "index '3'")
		assert.Contains(t, session.ErrorString(), "helpful text 3")
	})
}

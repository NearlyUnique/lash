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
	t.Run("where a default is set, missing env vars are set for the process", func(t *testing.T) {
		session := lash.NewSession()
		assert.Empty(t, os.Getenv("some-key"))
		session.
			Env().
			Default("some-key", "some-value")

		assert.Equal(t, "some-value", os.Getenv("some-key"))
	})
	t.Run("where a default is set, existing env vars are not overridden for the process", func(t *testing.T) {
		session := lash.NewSession()

		require.NoError(t, os.Setenv("some-key", "existing-value"))
		session.
			Env().
			Default("some-key", "some-value").
			Default("other-key", "other-value")

		assert.Equal(t, "existing-value", os.Getenv("some-key"))
		assert.Equal(t, "other-value", os.Getenv("other-key"))
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

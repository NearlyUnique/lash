package lash_test

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_can_append(t *testing.T) {
	t.Run("to new file", func(t *testing.T) {
		const aFilename = "aFilename.txt"
		_ = os.Remove(aFilename)
		defer func() { _ = os.Remove(aFilename) }()

		session := lash.NewSession()

		session.OpenFile(aFilename).AppendLine("a line")

		require.NoError(t, session.Err())

		actual, err := ioutil.ReadFile(aFilename)

		assert.NoError(t, err)
		assert.Equal(t, actual, []byte("a line\n"))
	})
	t.Run("to existing file", func(t *testing.T) {
		const aFilename = "aFilename.txt"

		require.NoError(t, ioutil.WriteFile(aFilename, []byte("line 1\n"), 0666))
		defer func() { _ = os.Remove(aFilename) }()

		session := lash.NewSession()

		session.OpenFile(aFilename).AppendLine("a line")

		require.NoError(t, session.Err())

		actual, err := ioutil.ReadFile(aFilename)

		assert.NoError(t, err)
		assert.Equal(t, string(actual), "line 1\na line\n")
	})
	t.Run("append line supports EnvStr", func(t *testing.T) {
		const aFilename = "aFilename.txt"

		require.NoError(t, ioutil.WriteFile(aFilename, []byte("line 1\n"), 0666))
		defer func() { _ = os.Remove(aFilename) }()

		session := lash.NewSession()
		require.NoError(t, os.Setenv("some_env", "some value"))

		session.OpenFile(aFilename).Truncate().AppendLine("$some_env $0", 99)

		require.NoError(t, session.Err())

		actual, err := ioutil.ReadFile(aFilename)

		assert.NoError(t, err)
		assert.Equal(t, "some value 99\n", string(actual))
	})
}

func Test_can_truncate_a_file(t *testing.T) {
	const aFilename = "aFilename.txt"
	require.NoError(t, ioutil.WriteFile(aFilename, []byte("line 1\n"), 0666))
	defer func() { _ = os.Remove(aFilename) }()

	session := lash.NewSession()
	session.
		OpenFile(aFilename).
		Truncate().
		AppendLine("some line")

	actual, err := ioutil.ReadFile(aFilename)
	assert.NoError(t, err)
	assert.Equal(t, string(actual), "some line\n")
}

func Test_can_append_concurrently_via_channel(t *testing.T) {
	const aFilename = "aFilename.txt"
	defer func() { _ = os.Remove(aFilename) }()

	session := lash.NewSession()

	appender := session.
		OpenFile(aFilename).
		Truncate().
		Appender()

	var wg sync.WaitGroup
	wg.Add(1)

	require.NoError(t, os.Setenv("some_env", "env value"))

	go func() {
		// both are equivalent
		appender.Ch() <- "any line"
		appender.AppendLine("last line $some_env ($0)", "a string")
		wg.Done()
	}()

	wg.Wait() // MUST finish YOUR work before you close the channel

	appender.Close()

	actual, err := ioutil.ReadFile(aFilename)
	assert.NoError(t, err)
	assert.Equal(t, "any line\nlast line env value (a string)\n", string(actual))
}

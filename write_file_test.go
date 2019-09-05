package lash_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

		scope := lash.NewScope()
		scope.OpenFile(aFilename).AppendLine("a line")

		require.NoError(t, scope.Err())

		actual, err := ioutil.ReadFile(aFilename)

		assert.NoError(t, err)
		assert.Equal(t, actual, []byte("a line\n"))
	})
	t.Run("to existing file", func(t *testing.T) {
		const aFilename = "aFilename.txt"

		require.NoError(t, ioutil.WriteFile(aFilename, []byte("line 1\n"), 0666))
		defer func() { _ = os.Remove(aFilename) }()

		scope := lash.NewScope()

		scope.OpenFile(aFilename).AppendLine("a line")

		require.NoError(t, scope.Err())

		actual, err := ioutil.ReadFile(aFilename)

		assert.NoError(t, err)
		assert.Equal(t, "line 1\na line\n", string(actual))
	})
	t.Run("append line supports EnvStr", func(t *testing.T) {
		aFilename := filepath.Join(os.TempDir(), "aFilename.txt")

		require.NoError(t, ioutil.WriteFile(aFilename, []byte("line 1\n"), 0666))
		defer func() { _ = os.Remove(aFilename) }()

		scope := lash.NewScope()
		require.NoError(t, os.Setenv("some_env", "some value"))

		scope.OpenFile(aFilename).Truncate().AppendLine("$some_env $0", 99)

		require.NoError(t, scope.Err())

		actual, err := ioutil.ReadFile(aFilename)

		assert.NoError(t, err)
		assert.Equal(t, "some value 99\n", string(actual))
	})
}

func Test_can_truncate_a_file(t *testing.T) {
	const aFilename = "aFilename.txt"
	require.NoError(t, ioutil.WriteFile(aFilename, []byte("line 1\n"), 0666))
	defer func() { _ = os.Remove(aFilename) }()

	scope := lash.NewScope()
	scope.
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

	scope := lash.NewScope()

	appender := scope.
		OpenFile(aFilename).
		Truncate().
		Appender()

	var wg sync.WaitGroup
	wg.Add(1)

	require.NoError(t, os.Setenv("some_env", "env value"))

	go func() {
		// both are equivalent
		appender.Ch() <- t.Name()
		appender.AppendLine("last line $some_env ($0)", "a string")
		wg.Done()
	}()

	wg.Wait() // MUST finish YOUR work before you close the channel

	appender.Close()

	actual, err := ioutil.ReadFile(aFilename)
	assert.NoError(t, err)
	assert.Equal(t, t.Name()+"\nlast line env value (a string)\n", string(actual))
	// try really hard to remove this test file
	_ = os.Remove(aFilename)
}

func Test_can_delete_a_file(t *testing.T) {
	filename := tempPathname()
	scope := lash.NewScope()
	scope.OpenFile(filename).AppendLine("any text").Close()

	_, err := os.Stat(filename)
	assert.NoError(t, err)

	scope.OpenFile(filename).Delete()

	_, err = os.Stat(filename)
	assert.True(t, os.IsNotExist(err))
}

func Test_creating_directories(t *testing.T) {
	aPath := tempPathname()
	scope := lash.NewScope()

	scope.OpenFile(aPath).Mkdir()

	assertDirExists(t, aPath)
}

func Test_when_copying_files_the_permissions_are_preserved(t *testing.T) {
	src, dest := tempPathname(), tempPathname()
	removeSrc := writeFile(t, src, "any-content")
	scope := lash.NewScope().OnError(requireNoError(t))

	scope.OpenFile(src).CopyTo(dest)

	assertFileExists(t, src)
	assertFileExists(t, dest)

	stat := func(filename string) os.FileMode {
		s, err := os.Stat(filename)
		require.NoError(t, err)
		return s.Mode()
	}
	assert.Equal(t, stat(src), stat(dest))

	removeSrc()
	os.Remove(dest)
}

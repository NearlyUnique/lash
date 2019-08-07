package lash_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_read_whole_file(t *testing.T) {
	const theContent = "any-content\nline2"
	const aFilename = "any-filename"

	require.NoError(t, ioutil.WriteFile(aFilename, []byte(theContent), 0777))
	defer func() { _ = os.Remove(aFilename) }()

	scope := lash.NewScope()
	actualContent := scope.
		OpenFile(aFilename).
		String()

	require.NoError(t, scope.Err())
	assert.Equal(t, theContent, actualContent)
}

func Test_open_file_supports_EnvStr(t *testing.T) {
	const theContent = "any-content"
	const aFilename = "any-filename"

	require.NoError(t, ioutil.WriteFile(aFilename, []byte(theContent), 0777))
	defer func() { _ = os.Remove(aFilename) }()

	require.NoError(t, os.Setenv("open_file_start", "any"))
	scope := lash.NewScope()

	actualContent := scope.
		OpenFile("$open_file_start-$0", "filename").
		String()

	require.NoError(t, scope.Err())
	assert.Equal(t, theContent, actualContent)
}

func Test_read_file_line_by_line(t *testing.T) {
	t.Run("read all lines when file is readable", func(t *testing.T) {

		const theContent = "line 1\nline 2\nline 3"
		const aFilename = "any-filename"

		require.NoError(t, ioutil.WriteFile(aFilename, []byte(theContent), 0777))
		defer func() { _ = os.Remove(aFilename) }()

		scope := lash.NewScope()
		ch := scope.
			OpenFile(aFilename).
			ReadLines()

		assert.Equal(t, "line 1", <-ch)
		assert.Equal(t, "line 2", <-ch)
		assert.Equal(t, "line 3", <-ch)
		_, ok := <-ch
		assert.False(t, ok)
	})
	t.Run("results in closed channel when error, error on the scope", func(t *testing.T) {
		const noSuchFile = "noSuchFile"
		scope := lash.NewScope()
		scope.OnError(lash.Ignore)
		ch := scope.
			OpenFile(noSuchFile).
			ReadLines()

		_, ok := <-ch
		assert.False(t, ok)
		assert.Contains(t, scope.Err().Error(), "File:ReadLines")
	})
}

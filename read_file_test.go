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

func Test_read_file_as_json(t *testing.T) {
	const aFilename = "any-filename"
	del := writeFile(t, aFilename, `{"name":"any name","count":99,"sub":{"ok":true}}`)
	defer del()

	var actual struct {
		Name  string
		Count int
		Sub   struct {
			OK bool
		}
	}

	scope := lash.NewScope()
	scope.OnError(requireNoError(t))
	scope.
		OpenFile(aFilename).
		AsJSON(&actual)

	assert.Equal(t, "any name", actual.Name)
	assert.Equal(t, 99, actual.Count)
	assert.True(t, actual.Sub.OK)
}

func Test_read_json_file_errors(t *testing.T) {
	t.Run("error when no such file", func(t *testing.T) {
		var actual struct{}
		var actualErr error
		scope := lash.NewScope()
		scope.OnError(func(e error) {
			actualErr = e
		})
		scope.
			OpenFile("no-such-file").
			AsJSON(&actual)

		assert.Error(t, actualErr)
		assert.Contains(t, actualErr.Error(), "no-such-file")
		assert.Contains(t, actualErr.Error(), "File:AsJSON_read")
	})
	t.Run("invalid json content", func(t *testing.T) {
		const aFilename = "any-filename"
		del := writeFile(t, aFilename, `not json }`)
		defer del()

		var actual struct{}
		var actualErr error
		scope := lash.NewScope()
		scope.OnError(func(e error) {
			actualErr = e
		})
		scope.
			OpenFile(aFilename).
			AsJSON(&actual)

		assert.Error(t, actualErr)
		assert.Contains(t, actualErr.Error(), aFilename)
		assert.Contains(t, actualErr.Error(), "File:AsJSON_unmarshal")
	})
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

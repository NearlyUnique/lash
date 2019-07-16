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

	session := lash.NewSession()
	actualContent := session.
		OpenRead(aFilename).
		String()

	assert.Equal(t, theContent, actualContent)
}

func Test_read_file_line_by_line(t *testing.T) {
	t.Run("read all lines when file is readable", func(t *testing.T) {

		const theContent = "line 1\nline 2\nline 3"
		const aFilename = "any-filename"

		require.NoError(t, ioutil.WriteFile(aFilename, []byte(theContent), 0777))
		defer func() { _ = os.Remove(aFilename) }()

		ch := lash.
			OpenRead(aFilename).
			ReadLines()

		assert.Equal(t, "line 1", <-ch)
		assert.Equal(t, "line 2", <-ch)
		assert.Equal(t, "line 3", <-ch)
		_, ok := <-ch
		assert.False(t, ok)
	})
	t.Run("results in closed channel when error, error on the session", func(t *testing.T) {
		const noSuchFile = "noSuchFile"
		defer lash.DefaultSession.ClearError()

		ch := lash.
			OpenRead(noSuchFile).
			ReadLines()

		_, ok := <-ch
		assert.False(t, ok)
		assert.Contains(t, lash.Error(), "File:ReadLines")
	})
}

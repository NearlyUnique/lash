package lash_test

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, name, content string) func() {
	require.NoError(t, ioutil.WriteFile(name, []byte(content), 0600))
	return func() { _ = os.Remove(name) }
}

func tempPathname() string {
	return filepath.Join(os.TempDir(), RndStr(10))
}

func assertFileExists(t *testing.T, filename string) {
	stat, err := os.Stat(filename)
	assert.NoError(t, err)
	assert.True(t, !stat.IsDir())
}

func assertDirExists(t *testing.T, filename string) {
	stat, err := os.Stat(filename)
	assert.NoError(t, err)
	assert.True(t, stat.IsDir())
}

// from https://www.calhoun.io/creating-random-strings-in-go/
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RndStr(length int) string {
	return stringWithCharset(length, charset)
}

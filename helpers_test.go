package lash_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, name, content string) func() {
	require.NoError(t, ioutil.WriteFile(name, []byte(content), 0600))
	return func() { _ = os.Remove(name) }
}

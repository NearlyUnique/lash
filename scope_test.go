package lash_test

import (
	"testing"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_simplify_making_json_data(t *testing.T) {
	val := struct {
		Text   string
		Number float32
	}{
		Text:   "some text",
		Number: 12.34,
	}

	scope := lash.NewScope()
	buf := scope.AsJson(val)
	require.NoError(t, scope.Err())
	assert.Equal(t, `{"Text":"some text","Number":12.34}`, string(buf))
}

func requireNoError(t *testing.T) func(error) {
	return func(err error) {
		require.NoError(t, err)
	}
}

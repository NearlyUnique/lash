package lash_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
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

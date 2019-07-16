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

	session := lash.NewSession()
	buf := session.AsJson(val)
	require.NoError(t, session.Err())
	assert.Equal(t, `{"Text":"some text","Number":12.34}`, string(buf))

	buf = lash.AsJson(val)
	require.NoError(t, lash.DefaultSession.Err())
	assert.Equal(t, `{"Text":"some text","Number":12.34}`, string(buf))
}

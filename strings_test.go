package lash_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/NearlyUnique/lash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_env_vars_can_be_expanded(t *testing.T) {
	keys := []string{"lower", "UPPER", "Mixed", "aZ_09"}
	for i, k := range keys {
		require.NoError(t, os.Setenv(k, fmt.Sprintf("value%d", i)))
	}
	require.NoError(t, os.Setenv("any-key", "some-value"))

	actual := lash.EnvStr("before $lower$UPPER ($Mixed) $aZ_09 after")

	assert.Equal(t, "before value0value1 (value2) value3 after", actual)
}
func Test_when_no_expansions_are_supplied_EnvStr_is_identity_function(t *testing.T) {
	assert.Equal(t, "no change", lash.EnvStr("no change"))
	assert.Equal(t, "", lash.EnvStr(""))
}
func Test_env_var_expansion_can_use_indexed_argument_values(t *testing.T) {
	a := someType{}
	actual := lash.EnvStr("$0, $1$2[$3]", 10, "any", true, a)

	assert.Equal(t, "10, anytrue[has-stringer]", actual)
}

type someType struct{}

func (someType) String() string { return "has-stringer" }

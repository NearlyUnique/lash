package lash_test

import (
	"bytes"
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

func Test_scope_version_caused_missing_values_to_generate_error(t *testing.T) {
	require.Empty(t, os.Getenv("no_such_env_var"))

	scope := lash.NewScope()
	scope.OnError(lash.Ignore)

	t.Run("missing arguments", func(t *testing.T) {
		scope.ClearError()
		actual := scope.EnvStr("causes error $0 $1", 1)

		assert.Error(t, scope.Err())
		assert.Equal(t, "causes error 1 ", actual)
		assert.Contains(t, scope.Err().Error(), "EnvStr:ArgIndex")
		assert.Contains(t, scope.Err().Error(), "'$1'")
	})
	t.Run("missing env vars", func(t *testing.T) {
		scope.ClearError()
		actual := scope.EnvStr("causes error $no_such_env_var")

		assert.Error(t, scope.Err())
		assert.Equal(t, "causes error ", actual)
		assert.Contains(t, scope.Err().Error(), "EnvStr:EnvName")
		assert.Contains(t, scope.Err().Error(), "'$no_such_env_var'")
	})
}

func Test_Println_allows_use_of_EnvStr(t *testing.T) {
	scope := lash.NewScope()
	actual := bytes.NewBuffer(nil)
	scope.SetOutput(actual)
	os.Setenv("some_value", "the value")
	scope.Println("any text $some_value here ($0)", 42)

	assert.Equal(t, "any text the value here (42)\n", actual.String())
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

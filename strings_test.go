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
func Test_session_version_caused_missing_values_to_generate_error(t *testing.T) {
	require.Empty(t, os.Getenv("no_such_env_var"))

	session := lash.NewSession()
	session.OnError(lash.Ignore)

	t.Run("missing arguments", func(t *testing.T) {
		session.ClearError()
		actual := session.EnvStr("causes error $0 $1", 1)

		assert.Error(t, session.Err())
		assert.Equal(t, "causes error 1 ", actual)
		assert.Contains(t, session.ErrorString(), "EnvStr:ArgIndex")
		assert.Contains(t, session.ErrorString(), "'$1'")
	})
	t.Run("missing env vars", func(t *testing.T) {
		session.ClearError()
		actual := session.EnvStr("causes error $no_such_env_var")

		assert.Error(t, session.Err())
		assert.Equal(t, "causes error ", actual)
		assert.Contains(t, session.ErrorString(), "EnvStr:EnvName")
		assert.Contains(t, session.ErrorString(), "'$no_such_env_var'")
	})
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

package lash_test

import (
	"os"
	"testing"

	"github.com/NearlyUnique/lash"
)

func Test_can_ensure_env_has_value(t *testing.T) {
	_ = os.Setenv("some-key", "some-value")

	session := lash.NewSession()

	session.
		Env().
		Require("some-key", "some description")
}

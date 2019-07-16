package lash

import (
	"flag"
	"fmt"
	"os"
)

type (
	RequireEnv struct {
		session *Session
	}
	RequireFlag struct {
		session *Session
		flagSet flag.FlagSet
	}
)

func (r RequireEnv) Require(key, description string) {
	if v := os.Getenv(key); len(v) == 0 {
		r.session.SetErr(&SessionErr{
			Type:   "Env",
			Action: "Require",
			Err:    fmt.Errorf("missing '%s': %s", key, description),
		})
	}
}

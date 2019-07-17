package lash

import (
	"fmt"
	"os"
)

type (
	// RequireEnv vars to the program
	RequireEnv struct {
		session *Session
	}
	// RequireFlag set to teh program
	//RequireFlag struct {
	//	session *Session
	//	Flags   flag.FlagSet
	//}
	// RequireArgs to the program
	RequireArg struct {
		session *Session
		Args    []string
	}
)

// Require that en env var has a value, error contains the description
func (r RequireEnv) Require(key, description string) {
	if v := os.Getenv(key); len(v) == 0 {
		r.session.SetErr(&SessionErr{
			Type:   "Env",
			Action: "Require",
			Err:    fmt.Errorf("missing '%s': %s", key, description),
		})
	}
}

// Require tests args as per os.Args 0th arg is program name
func (r RequireArg) Require(index int, description string) RequireArg {
	if len(r.Args)-1 < index {
		r.session.SetErr(&SessionErr{
			Type:   "Arg",
			Action: "Require",
			Err:    fmt.Errorf("missing index '%d': %s", index, description),
		})
	}
	return r
}

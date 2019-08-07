package lash

import (
	"fmt"
	"os"
)

type (
	// RequireEnv vars to the program
	RequireEnv struct {
		scope *Scope
	}
	// RequireFlag set to teh program
	//RequireFlag struct {
	//	scope *Scope
	//	Flags   flag.FlagSet
	//}
	// RequireArgs to the program
	RequireArg struct {
		scope *Scope
		Args  []string
	}
)

// Require that en env var has a value, error contains the description
func (r RequireEnv) Require(key, description string) RequireEnv {
	if v := os.Getenv(key); len(v) == 0 {
		r.scope.SetErr(&ScopeErr{
			Type:   "Env",
			Action: "Require",
			Err:    fmt.Errorf("missing '%s': %s", key, description),
		})
	}
	return r
}

func (r RequireEnv) Default(key, defaultValue string) RequireEnv {
	if os.Getenv(key) == "" {
		r.scope.SetErr(&ScopeErr{
			Type:   "Env",
			Action: "SetDefault",
			Err:    os.Setenv(key, defaultValue),
		})
	}
	return r
}

// Require tests args as per os.Args 0th arg is program name
func (r RequireArg) Require(index int, description string) RequireArg {
	if len(r.Args)-1 < index {
		r.scope.SetErr(&ScopeErr{
			Type:   "Arg",
			Action: "Require",
			Err:    fmt.Errorf("missing index '%d': %s", index, description),
		})
	}
	return r
}

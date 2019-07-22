package lash

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"golang.org/x/xerrors"
)

var rxDollarEnv = regexp.MustCompile(`\$(?P<Key>[a-zA-Z0-9_]+)`)

func EnvStr(msg string, args ...interface{}) string {
	return replaceAllStringSubMatchFunc(rxDollarEnv, msg, func(values []string) string {
		if i, err := strconv.Atoi(values[1]); err == nil && i < len(args) {
			return fmt.Sprintf("%v", args[i])
		}
		return os.Getenv(values[1])
	})
}

//EnvStr as lash.EnvStr but causes error if format is malformed or sources are missing
func (s *Session) EnvStr(msg string, args ...interface{}) string {
	serr := SessionErr{Type: "EnvStr"}
	return replaceAllStringSubMatchFunc(rxDollarEnv, msg, func(values []string) string {
		if i, err := strconv.Atoi(values[1]); err == nil {
			if i < 0 || i >= len(args) {
				s.SetErr(serr.fail("ArgIndex", xerrors.Errorf("'$%d' is out of range in '%s'", i, msg)))
				return ""
			}
			return fmt.Sprintf("%v", args[i])
		}
		v := os.Getenv(values[1])
		if v == "" {
			s.SetErr(serr.fail("EnvName", xerrors.Errorf("'$%s' not found in '%s'", values[1], msg)))
		}
		return v
	})
}

func replaceAllStringSubMatchFunc(re *regexp.Regexp, str string, repl func(args []string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		var groups []string
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

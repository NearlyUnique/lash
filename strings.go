package lash

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
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

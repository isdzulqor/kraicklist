package errors

import (
	"fmt"
	"strings"
)

func FormatQueryError(query string, args ...interface{}) (out string) {
	for i, arg := range args {
		toBeReplaced := fmt.Sprintf("$%d", i+1)
		out = strings.Replace(query, toBeReplaced, anyToQueryParam(arg), -1)
		query = out
	}
	if out == "" {
		return fmt.Sprintf("query: %s\targs: %s", query, args)
	}
	return
}

func anyToQueryParam(in interface{}) (out string) {
	switch v := in.(type) {
	case string:
		out = "'" + fmt.Sprint(v) + "'"
		return
	}
	out = fmt.Sprint(in)
	if out == "<nil>" {
		out = "null"
	}
	return
}

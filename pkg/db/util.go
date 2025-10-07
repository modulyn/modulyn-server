package db

import (
	"fmt"
	"strings"
)

func interpolateSQL(query string, args ...any) string {
	var b strings.Builder
	argIdx := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' && argIdx < len(args) {
			switch v := args[argIdx].(type) {
			case string:
				b.WriteString(fmt.Sprintf("'%s'", v))
			case []byte:
				b.WriteString(fmt.Sprintf("'%x'", v))
			case nil:
				b.WriteString("NULL")
			case bool:
				if v {
					b.WriteString("1")
				} else {
					b.WriteString("0")
				}
			default:
				b.WriteString(fmt.Sprintf("%v", v))
			}
			argIdx++
		} else {
			b.WriteByte(query[i])
		}
	}
	return b.String()
}

package utility

import (
	"strings"
)

func GetLastSplit(src, separator string) string {
	split := strings.Split(src, separator)
	return split[len(split)-1]
}

package utils

import (
	"strings"
)

func GenerateKey(keys ...string) string {
	return strings.Join(keys, ":")
}

package utils

import (
	"fmt"
	"strings"
)

func GenerateKey(clientId string, keys ...string) string {
	return fmt.Sprintf("%s:%v", clientId, strings.Join(keys, ":"))
}

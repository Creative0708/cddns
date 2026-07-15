package main

import (
	"fmt"
	"strings"
)

const APIKEY_PREFIX = "CDDNS_"
const APIKEY_MIN_LEN = 24

type ApiKeyError string

func (e ApiKeyError) Error() string {
	return string(e)
}

func validateApiKey(key string) error {
	if !strings.HasPrefix(key, APIKEY_PREFIX) {
		return ApiKeyError("string does not start with " + APIKEY_PREFIX)
	}
	trimmed := []byte(key)[len(APIKEY_PREFIX):]
	if len(trimmed) < APIKEY_MIN_LEN {
		return ApiKeyError(fmt.Sprintf("invalid api key length %d", len(trimmed)))
	}
	for _, ch := range trimmed {
		if !(ch >= 'A' && ch <= 'Z' || ch >= '2' && ch <= '7') {
			return ApiKeyError(fmt.Sprintf("invalid char '%c' in api key", ch))
		}
	}
	return nil
}

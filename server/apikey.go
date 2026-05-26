package main

import (
	"crypto/rand"
)

const APIKEY_PREFIX = "CDDNS_"
const APIKEY_LEN = 64

func generateApiKey() string {
	return APIKEY_PREFIX + rand.Text()
}

package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateUUID generates an uuid
// https://tools.ietf.org/html/rfc4122
// crypto.rand use getrandom(2) or /dev/urandom
// It is maybe occur an error due to system error
// panic if an error occurred
func GenerateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		panic("generate an uuid failed, error: " + err.Error())
	}
	// see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

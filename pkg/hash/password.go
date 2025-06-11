package hash

import (
	"crypto/sha1"
	"fmt"
)

const saltPASSWORD = "ljnks89t7y32y9jsdfjkh1289"

func PasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(saltPASSWORD))
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

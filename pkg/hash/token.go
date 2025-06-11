package hash

import (
	"crypto/sha1"
	"fmt"
)

const saltJWT = "JKFI2WEF48FEu9hdfRTasjdlSDFGKL3yr9KH76FGJdkf"

func Token(jwtToken string) string {
	hash := sha1.New()
	hash.Write([]byte(jwtToken))
	hash.Write([]byte(saltJWT))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

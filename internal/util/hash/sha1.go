package hash

import (
	"crypto/sha1"
	"encoding/hex"
)

func Sha1(s string) string {
	rs := sha1.New()

	rs.Write([]byte(s))

	bts := rs.Sum(nil)

	return hex.EncodeToString(bts)
}

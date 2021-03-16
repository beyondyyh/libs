package conhash

import (
	"crypto/md5"
)

// use successive 4-bytes from hash as numbers
func hashDef(key string) uint64 {
	md5sum := md5.New()
	md5sum.Write([]byte(key))
	cstr := md5sum.Sum(nil)

	var hash uint64
	for i := 0; i < 4; i++ {
		hash += ((uint64)(cstr[i*4+3]&0xFF) << 24) | ((uint64)(cstr[i*4+2]&0xFF) << 16) | ((uint64)(cstr[i*4+1]&0xFF) << 8) | ((uint64)(cstr[i*4+0] & 0xFF))
	}

	return hash
}

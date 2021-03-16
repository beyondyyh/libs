package openssl

import (
	"crypto/des"
)

// DesECBEncrypt
// Des算法，ECB应用模式加密
func DesECBEncrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return ECBEncrypt(block, src, padding)
}

// DesECBDecrypt
// Des算法，ECB应用模式解密
func DesECBDecrypt(src, key []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return ECBDecrypt(block, src, padding)
}

// DesCBCEncrypt
// Des算法，CBC应用模式加密
func DesCBCEncrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCEncrypt(block, src, iv, padding)
}

// DesCBCDecrypt
// Des算法，CBC应用模式解密
func DesCBCDecrypt(src, key, iv []byte, padding string) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return CBCDecrypt(block, src, iv, padding)
}

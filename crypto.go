package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
)

var salt = []byte{89, 49, 108, 82, 52, 109, 67, 106, 89, 104, 80, 112, 100, 64, 55, 48, 78, 121, 48, 69}

func hashPassword(p []byte) []byte {
	h := sha256.New()
	h.Write(append(p, salt...))
	return h.Sum(nil)
}

func decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func decrypt(key []byte, b64 string) []byte {
	text := decodeBase64(b64)
	block, err := aes.NewCipher(hashPassword(key))
	if err != nil {
		panic(err)
	}
	if len(text) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	return text
}

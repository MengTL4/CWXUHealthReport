package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func pkcs7padding(text string) string {
	bs := 16
	bytesLength := len([]byte(text))
	paddingSize := bs - bytesLength%bs
	paddingText := bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)
	return text + string(paddingText)
}

func encryptByAES(message string) string {
	keyword := "u2oh6Vu^HWe4_AES"
	key := []byte(keyword)
	iv := []byte(keyword)

	block, _ := aes.NewCipher(key)
	contentPadding := pkcs7padding(message)
	ciphertext := make([]byte, len(contentPadding))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, []byte(contentPadding))

	result := base64.StdEncoding.EncodeToString(ciphertext)
	return result
}

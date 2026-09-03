package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func GenerateAESEncryptionSettings() ([]byte, []byte, error) {
	const charset = "0123456789"

	key := make([]byte, 16) // AES-128
	iv := make([]byte, 16)  // AES block size

	for i := range key {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, nil, err
		}
		key[i] = charset[num.Int64()]
	}

	for i := range iv {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, nil, err
		}
		iv[i] = charset[num.Int64()]
	}

	return key, iv, nil
}

func EncryptRSAOAEP(modulus, exponent, message []byte) (string, error) {
	publicKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulus),
		E: int(new(big.Int).SetBytes(exponent).Int64()),
	}

	// SHA-1, not SHA-256: the router's OAEP decryption matches PyCryptodome's
	// PKCS1_OAEP default hash.
	cipherBytes, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, message, nil)
	if err != nil {
		return "", err
	}

	hexCiphertext := hex.EncodeToString(cipherBytes)

	for len(hexCiphertext) < len(modulus)*2 {
		hexCiphertext = "0" + hexCiphertext
	}

	return hexCiphertext, nil
}

func EncryptPKCS1v15(modulus, exponent, message []byte) (string, error) {
	publicKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulus),
		E: int(new(big.Int).SetBytes(exponent).Int64()),
	}

	cipherBytes, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, message)
	if err != nil {
		return "", err
	}

	hexCiphertext := hex.EncodeToString(cipherBytes)

	for len(hexCiphertext) < len(modulus)*2 {
		hexCiphertext = "0" + hexCiphertext
	}

	return hexCiphertext, nil
}

func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	paddedBytes := make([]byte, padding)

	for i := range paddedBytes {
		paddedBytes[i] = byte(padding)
	}

	return append(data, paddedBytes...)
}

func PKCS7Unpadding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("Empty PKCS#7 byte array")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize {
		return nil, fmt.Errorf("Invalid PKCS#7 padding")
	}

	for _, b := range data[len(data)-padding:] {
		if int(b) != padding {
			return nil, fmt.Errorf("PKCS#7 padding mismatch")
		}
	}

	return data[:len(data)-padding], nil
}

func EncryptAES128CBCPKCS7(key, iv []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	paddedBytes := PKCS7Padding(data, aes.BlockSize)

	encryptedData := make([]byte, len(paddedBytes))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedData, paddedBytes)

	return encryptedData, nil
}

func DecryptAES128CBCPKCS7(key, iv []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decryptedData := make([]byte, len(data))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decryptedData, data)

	return PKCS7Unpadding(decryptedData)
}

func HMACSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)

	return mac.Sum(nil)
}

package relaycrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"unsafe"

	"github.com/go-faster/xor"
)

const NonceSize int = 6

func XOREncryptIP(ip []byte, key []byte, dst []byte) int {
	// this is a really bad idea - (a^k)^(b^k)=a^b
	return xor.Bytes(dst, ip, key)
}

func XORDecryptIP(ciphertext []byte, key []byte, dst []byte) int {
	return xor.Bytes(dst, ciphertext, key)
}

type AESEncryptor struct {
	aesgcm cipher.AEAD
	NextNonce []byte
}

func MakeAES128Encryptor(key []byte) *AESEncryptor {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Failed to make aes128 block, ", err)
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, NonceSize)
	if err != nil {
		fmt.Println("Failed to make gcm from aes128 block, ", err)
	}
	return &AESEncryptor{aesgcm, make([]byte, NonceSize),}
}

func (encryptor *AESEncryptor) AES128EncryptIP(ip []byte, dst []byte) []byte {
	ret := encryptor.aesgcm.Seal(dst[:0], encryptor.NextNonce, ip, nil)

	// this is around 3x slower than a direct addition on uint64, which is negligible
	// memory error if we run out of 6 byte nonces, add key-renegotiation later
	*(*uint64)(unsafe.Pointer(unsafe.SliceData(encryptor.NextNonce))) += 1
	return ret
}

type AESDecryptor struct {
	aesgcm cipher.AEAD
}

func MakeAES128Decryptor(key []byte) *AESDecryptor {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Failed to make a decryptor, ", err)
		return nil
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, NonceSize)
	if err != nil {
		fmt.Println("Failed to make gcm in decryptor init, ", err)
		return nil
	}
	return &AESDecryptor{aesgcm: aesgcm}
}
func (decryptor *AESDecryptor) AES128DecryptIP(ciphertext []byte, nonce []byte, dst []byte) ([]byte, error) {
	b, err := decryptor.aesgcm.Open(dst[0:], nonce, ciphertext, nil)
	return b, err
}



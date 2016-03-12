package utility

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	log "github.com/cihub/seelog"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

type BinaryToText func(data []byte) string
type TextToBinary func(data string) ([]byte, error)

// Sha1SumWithSalt caculates SHA1 hash with salt
func Sha1SumWithSalt(seed, salt string) string {
	input := seed + salt

	result := fmt.Sprintf("%x", sha1.Sum([]byte(input)))

	log.Debugf("Sha1Sum: seed[%s], salt[%s], result[%s]", seed, salt, result)

	return result
}

func Aes128Encrypt(key, cipherString, src string, encode BinaryToText) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Debugf("Fail NewCipher: err[%s]", err.Error())

		return "", err
	}

	str := []byte(src)

	ciphertext := []byte(cipherString)
	iv := ciphertext[:aes.BlockSize]
	encrypter := cipher.NewCFBEncrypter(block, iv)

	encrypted := make([]byte, len(str))
	encrypter.XORKeyStream(encrypted, str)

	result := encode(encrypted)

	return result, nil
}

func Aes128Decrypt(key, cipherString, encodedSrc string, decode TextToBinary) (string, error) {
	encrypted, err := decode(encodedSrc)
	if err != nil {
		log.Debugf("Fail base32 decode: err[%s]", err.Error())

		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Debugf("Fail NewCipher: err[%s]", err.Error())

		return "", err
	}

	ciphertext := []byte(cipherString)
	iv := ciphertext[:aes.BlockSize]
	decrypter := cipher.NewCFBDecrypter(block, iv)

	decrypted := make([]byte, len(encrypted))
	decrypter.XORKeyStream(decrypted, encrypted)

	return string(decrypted[:]), nil
}

func Base64CustomeAlphabetEncode(alphabet string, src []byte) string {
	encoding := base64.NewEncoding(alphabet)

	return encoding.EncodeToString(src)
}

func Base64CustomeAlphabetDecode(alphabet string, base64Src string) ([]byte, error) {
	encoding := base64.NewEncoding(alphabet)

	return encoding.DecodeString(base64Src)
}

func IsBase32Encoded(data string) bool {
	_, err := base32.StdEncoding.DecodeString(data)
	if err != nil {
		return false
	}
	return true
}

// RandSeq generates the random string of lenght n
func RandSeq(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// CuteBi_encrypt.go
package main

import (
	"encoding/base64"
	"errors"
	"time"
)

var (
	passStr_code byte
	passLen_code byte
)

func CuteBi_encrypt(data []byte, dynamic_code byte) byte {
	for i := 0; i < len(data); i++ {
		data[i] = (data[i] - passLen_code) ^ dynamic_code
		data[i] = ((data[i]<<(dynamic_code%8) | data[i]>>(8-(dynamic_code%8))) + dynamic_code) ^ passStr_code
		dynamic_code += (dynamic_code & passStr_code) ^ (data[i] | passLen_code)
	}

	return dynamic_code
}

func CuteBi_decrypt(data []byte, dynamic_code byte) byte {
	var en_byte byte

	for i := 0; i < len(data); i++ {
		en_byte = data[i]
		data[i] = (data[i] ^ passStr_code) - dynamic_code
		data[i] = ((data[i]>>(dynamic_code%8) | data[i]<<(8-(dynamic_code%8))) ^ dynamic_code) + passLen_code
		dynamic_code += (dynamic_code & passStr_code) ^ (en_byte | passLen_code)
	}

	return dynamic_code
}

func make_CuteBi_encrypt_PassCode(password []byte) {
	var i, password_len byte

	password_len = byte(len(password))
	passStr_code = 0
	for i = 0; i < password_len; i++ {
		if password[i]&byte(1) == 1 {
			passStr_code += password[i] | i
			passStr_code -= password_len
		} else {
			passStr_code *= password_len
			passStr_code |= password[i] & i
		}
	}
	if password_len > 5 && passStr_code != 0 {
		passLen_code = password_len * passStr_code
	} else {
		passLen_code = password_len << 4
	}
}

func make_CuteBiEncrypt_dynamic_code(timeIgnoreTenBit uint32) byte {
	return byte((((((timeIgnoreTenBit) << 30) >> 26) | ((((timeIgnoreTenBit) >> 2) << 30) >> 24) | ((((timeIgnoreTenBit) >> 7) << 29) >> 29)) | ((((timeIgnoreTenBit) >> 4) << 28) >> 29)) + ((((timeIgnoreTenBit) << 28) >> 28) & ((timeIgnoreTenBit) >> 1)))
}

func verify_CuteBiEncrypt_dynamic_code(data []byte) (byte, bool) {
	tmpData := make([]byte, len(data))
	timeIgnore7Bit := uint32(time.Now().Unix() >> 7)
	decrypt_dynamic_code := make_CuteBiEncrypt_dynamic_code(timeIgnore7Bit)

	copy(tmpData, data)
	CuteBi_decrypt(tmpData, decrypt_dynamic_code)
	if tmpData[len(tmpData)-1] != 0 {
		decrypt_dynamic_code = make_CuteBiEncrypt_dynamic_code(timeIgnore7Bit - 1)
		copy(tmpData, data)
		CuteBi_decrypt(tmpData, decrypt_dynamic_code)
		if tmpData[len(tmpData)-1] != 0 {
			decrypt_dynamic_code = make_CuteBiEncrypt_dynamic_code(timeIgnore7Bit + 1)
			copy(tmpData, data)
			CuteBi_decrypt(tmpData, decrypt_dynamic_code)
			if tmpData[len(tmpData)-1] != 0 {
				return 0, false
			}
		}
	}
	return decrypt_dynamic_code, true
}

func CuteBi_decrypt_host(host []byte) ([]byte, byte, error) {
	hostDec := make([]byte, len(host))
	n, err := base64.StdEncoding.Decode(hostDec, host)
	if err != nil {
		return nil, 0, err
	}
	dynamic_code, decryptSuccess := verify_CuteBiEncrypt_dynamic_code(hostDec[:n])
	if decryptSuccess == false {
		return nil, 0, errors.New("Error host decrypt")
	}
	CuteBi_decrypt(hostDec[:n-1], dynamic_code)

	return hostDec[:n-1], dynamic_code, nil
}

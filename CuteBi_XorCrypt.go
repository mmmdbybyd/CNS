// CuteBi_XorCrypt.go
package main

import (
	"encoding/base64"
	"errors"
)

var CuteBi_XorCrypt_password []byte

/* 一个简单的异或加密 */
func CuteBi_XorCrypt(data []byte, passwordSub int) int {
	for dataSub := 0; dataSub < len(data); {
		data[dataSub] ^= CuteBi_XorCrypt_password[passwordSub] | byte(passwordSub) //如果只是data[dataSub] ^= CuteBi_XorCrypt_password[passwordSub]，则密码"12"跟密码"1212"没用任何区别
		dataSub++
		passwordSub++
		if passwordSub == len(CuteBi_XorCrypt_password) {
			passwordSub = 0
		}
	}

	return passwordSub
}

func CuteBi_decrypt_host(host []byte) ([]byte, error) {
	hostDec := make([]byte, len(host))
	n, err := base64.StdEncoding.Decode(hostDec, host)
	if err != nil {
		return nil, err
	}
	CuteBi_XorCrypt(hostDec, 0)
	if hostDec[n-1] != 0 {
		return nil, errors.New("Decrypt failed.")
	}

	return hostDec[:n-1], nil
}

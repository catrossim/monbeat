package utils

import (
	"crypto/md5"
	"fmt"
)

func GenFileToken(token []byte) (string, error) {
	m := md5.New()
	_, err := m.Write(token)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", m.Sum(nil)), nil
}

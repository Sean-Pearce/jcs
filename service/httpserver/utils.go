package main

import (
	"crypto/rand"
	"encoding/base64"
)

func genToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func getUsernameByToken(token string) string {
	username := tokenMap[token]
	return username
}

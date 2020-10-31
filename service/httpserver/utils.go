package main

import (
	"encoding/base64"
	"math/rand"
	"time"
)

func genToken() string {
	b := make([]byte, 8)
	rand.Seed(time.Now().UnixNano())
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func genAccessKey() string {
	b := make([]byte, 12)
	rand.Seed(time.Now().UnixNano())
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func getUsernameByToken(token string) string {
	username := tokenMap[token]
	return username
}

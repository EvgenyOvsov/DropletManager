package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"
)

type DeleteRequest struct {
	ID string `json:"id"`
	Expiration string `json:"expiration"`
}

type CreateRequest struct {
	Key string `json:"key"`
	Type string `json:"type"`
	User struct{
		Email string `json:"email"`
		Name string `json:"name"`
	} `json:"user"`
	Trial bool `json:"trial"`
	Download string `json:"download"`
}

type UpdateRequest struct {
	Key string `json:"key"`
	IP string `json:"ip"`
}

type EmailRequest struct {
	Token string `json:"token"`
	From string `json:"from"`
	To []string `json:"to"`
	Subject string `json:"subject"`
	Body string `json:"body"`
}

func generateSHA() string{
	h := sha256.New()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	h.Write([]byte(fmt.Sprintf("%v",r.Intn(1000000))))
	return fmt.Sprintf("%x", h.Sum(nil))
}
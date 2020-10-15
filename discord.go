/*
	More about this part of code you can find at...
	https://vicuesoft.ru:2443/microservices/discord
	https://github.com/EvgenyOvsov/DiscordMicroservice
*/

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Notify(text string){
	r := struct {
		Token string `json:"token"`
		Channel string `json:"channel"`
		Text string `json:"text"`
		Color string `json:"color"`
	}{
		Token: "0[DELETED]f",
		Channel: "portal",
		Text: text,
		Color: "green",
	}
	j, _ := json.Marshal(r)
	reader := bytes.NewReader(j)
	http.Post("http://vicuesoft.com:5001", "application/json", reader)
}

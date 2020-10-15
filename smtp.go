/*
	More about this part of code you can find at...
	https://vicuesoft.ru:2443/microservices/mailmaster
	https://github.com/EvgenyOvsov/Mailer
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func(e *EmailRequest)Send(){
	j, _ := json.Marshal(e)
	reader := bytes.NewReader(j)
	resp, err := http.Post("http://vicuesoft.com:5000", "application/json", reader)
	if err != nil {
		Notify("Дроплет Менеджер: Я не смог отправить письмо.\n"+err.Error())
		return
	}
	if resp.StatusCode!=200{
		time.Sleep(time.Second*10)
		http.Post("http://vicuesoft.com:5000", "application/json", reader)
	}
}

func FlexLetter(request CreateRequest, ip string){
	l := &EmailRequest{
		Token:   "0[DELETED]ff",
		From:    "noreply@vicuesoft.com",
		To:      []string{request.User.Email},
		Subject: "Important information on your ViCue Soft products",
		Body:    fmt.Sprintf(
			"Dear %s,\r\n\r\n" +
				"Thank you for purchasing VQ Analyzer Subscription! Please retain this information for future reference.\r\n\r\n" +
				"Your credentials (IP address & port) to activate: %v:7000\r\n\r\n" +
				"Your License Key to cancel Sub: %v\r\n\r\n"+
				"Your download link (valid next 48 hours): %v\r\n\r\n" +
				"To unsubscribe go to: https://portal.vicuesoft.com/cancel \r\n\r\n" +
				"Please read FAQ if you got any questions: https://vicuesoft.com/vqasubfaq \r\n\r\n" +
				"With Best Regards,\r\n\r\n" +
				"ViCue Soft",
			request.User.Name,
			ip,
			request.Key,
			request.Download),
	}
	l.Send()
}
func AlexeyFadeevNotification(request CreateRequest, ip string){
	l := &EmailRequest{
		Token:   "0x[DELETED]f",
		From:    "noreply@vicuesoft.com",
		To:      []string{"alexey.fadeev@vicuesoft.com"},
		Subject: "[Portal] New license registered",
		Body: fmt.Sprintf(
			"Dear Alexey,\r\n\r\n"+
				"%v got license via Portal...\r\n\r\n"+
				"License Key to activate: %v\r\n\r\n"+
				"Type of product: VQAnalyzer %v FLEX\r\n\r\n"+
				"Floating server created: %v :7000\r\n\r\n"+
				"With Best Regards,\r\n\r\n"+
				"ViCue Soft",
			request.User.Email,
			request.Key,
			request.Type,
			ip),
	}
	l.Send()
}

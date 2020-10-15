package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)
const(
	password = "yC[DELETED]ouu"
)
var (
	CTX = context.Background()
	RDB *redis.Client
	INTERVAL = time.Minute*5
)
func RedisInit(){
	RDB = redis.NewClient(&redis.Options{
		Addr:     "portal.vicuesoft.com:6379",
		Password: password,
		DB:       1,
		MaxConnAge: time.Hour*24,
	})

	renew := func(c *redis.Client){
		for true{
			time.Sleep(time.Hour*24)
			RDB = redis.NewClient(&redis.Options{
				Addr:     "portal.vicuesoft.com:6379",
				Password: password,
				DB:       1,
				MaxConnAge: time.Hour*24,
			})
		}
	}
	go renew(RDB)
}

func AddDelete(droplet DeleteRequest){
	j, err := json.Marshal(droplet)
	if err != nil {
		log.Println(err)
		return
	}
	expiration, _ := time.Parse(time.ANSIC, droplet.Expiration)
	RDB.Set(CTX, droplet.ID, j, expiration.Sub(time.Now()))
	RDB.SAdd(CTX, "Timers", droplet.ID)
}

func GCheckLoop(){
	ExitConditions.Add(1)
	defer ExitConditions.Done()
	for true{
		list, err := RDB.SMembers(CTX, "Timers").Result()
		if err != nil {
			log.Println(err)
		}
		for _, v := range list{
			_, err := RDB.Get(CTX, v).Result()
			if err != nil {
				log.Printf("Found burned droplet! %v\n", v)
				DeleteDroplet(v)
				RDB.SRem(CTX, "Timers", v)
				continue
			}
		}
		time.Sleep(INTERVAL)
	}
}

func GDeleteListener(){
	ExitConditions.Add(1)
	defer ExitConditions.Done()
	handle := func(message string){
		var m DeleteRequest
		err := json.Unmarshal([]byte(message), &m)
		if err != nil {
			return
		}
		if _, err = time.Parse(time.ANSIC, m.Expiration); err!=nil{
			return
		}
		AddDelete(m)
	}
	for true{
		sub := RDB.Subscribe(CTX, "DeleteChannel")
		_, err := sub.ReceiveMessage(CTX)
		if err != nil {continue}
		message, err := RDB.LPop(CTX, "DeleteQueue").Result()
		if err != nil {continue}
		go handle(message)
	}
}

func GCreateListener(){
	ExitConditions.Add(1)
	defer ExitConditions.Done()
	handle := func(message string){
		var m CreateRequest
		log.Println(message)
		err := json.Unmarshal([]byte(message), &m)
		if err != nil {
			fmt.Println("Cant parse")
			return
		}
		if m.Key=="" || m.Type=="" || m.User.Email == ""{
			log.Println("Cant parse")
			return
		}
		CreateDroplet(m)
	}
	for true{
		sub := RDB.Subscribe(CTX, "CreateChannel")
		_, err := sub.ReceiveMessage(CTX)
		if err != nil {continue}
		message, err := RDB.LPop(CTX, "CreateQueue").Result()
		if err != nil {continue}
		go handle(message)
	}
}
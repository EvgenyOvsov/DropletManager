package main

import (
	"sync"
	"time"
)

var ExitConditions sync.WaitGroup

func main(){
	RedisInit()
	DigitalOceanInit()
	go GCheckLoop()
	go GDeleteListener()
	go GCreateListener()
	time.Sleep(20*time.Millisecond)
	ExitConditions.Wait()
}

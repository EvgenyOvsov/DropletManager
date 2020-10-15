package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/digitalocean/godo"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const oceanToken = "71a5a17b128622c7e5[DELETED]9ed6bc60b6cdea952c7d2"
var DOClient *godo.Client

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func DigitalOceanInit(){
	tokenSource := &TokenSource{
		AccessToken: oceanToken,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	DOClient = godo.NewClient(oauthClient)

	renew := func(client *godo.Client){
		for true{
			time.Sleep(time.Minute*5)
			client = godo.NewClient(oauthClient)
		}
	}
	go renew(DOClient)
}

func DeleteDroplet(id string){
	log.Println("DeleteRequest ", id, " will be destroyed!")
	if !isactive(id){RDB.SRem(CTX, "timers", id);return}
	i, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}
	_, err = DOClient.Droplets.Delete(CTX, i)
	if err != nil {
		Notify("Дроплет "+id+" должен быть уничтожен, но менеджер не смог этого сделать!")
		log.Println(err)
		return
	}
	Notify("Дроплет "+id+" уничтожен, потому что его время кончилось.")
}

func CreateDroplet(request CreateRequest){
	createRequest := &godo.DropletCreateRequest{
		Name:   fmt.Sprintf("ubuntu-SF-auto-%s-%s-%s",
			request.Type,
			strings.Split(request.User.Email, "@")[0],
			generateSHA()[:5]),
		Region: "sfo2",
		Size:   "s-1vcpu-1gb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-18-04-x64",
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			godo.DropletCreateSSHKey{ID: 25991323},
			godo.DropletCreateSSHKey{ID: 26686493},
			godo.DropletCreateSSHKey{ID: 26922516},
		},
		IPv6: false,
		Tags: []string{"AUTO"},
	}
	droplet, _, err := DOClient.Droplets.Create(CTX, createRequest)
	if err != nil {
		Notify(fmt.Sprintf("Менеджер не смог создать сервер.\n%v", err))
		return
	}
	time.Sleep(time.Minute*2)
	droplet, _, err = DOClient.Droplets.Get(CTX, droplet.ID)
	if err != nil {
		Notify(fmt.Sprintf("Менеджер не смог создать сервер.\n%v", err))
		return
	}
	counter := 4
	for true{
		droplet, _, err = DOClient.Droplets.Get(CTX, droplet.ID)
		if err != nil {
			Notify(fmt.Sprintf("Менеджер не смог создать сервер.\n%v", err))
			return
		}
		if droplet.Status == "new"{
			counter += 1
			if counter>10{
				return
			}
			time.Sleep(time.Second*30)
		}else{
			break
		}
	}
	ip, _ := droplet.PublicIPv4()
	for _,v := range []string{
		fmt.Sprintf("echo \"%s\" > /opt/.key", request.Key),
		"apt update",
		"wget -P /opt/ https://vicue[DELETED].sh",
		"chmod +x /opt/[DELETED].sh",
		fmt.Sprintf("/opt/[DELETED].sh %s %s",request.Key, request.Type),
	}{
		err = runCommand(ip, v)
		if err != nil {
			Notify(fmt.Sprintf("В процессе создания сервера %v для %v что-то пошло не так.\n%v", ip, request.User.Email, err))
			return
		}
	}
	time.Sleep(10*time.Second)
	if floatingCheck(ip){
		Notify(fmt.Sprintf("Был создан новый облачный сервер с адресом **%v** для **%v**", ip, request.User.Email))
		j, _ := json.Marshal(UpdateRequest{
			Key: request.Key,
			IP:  ip,
		})
		RDB.LPush(CTX, "UpdateQueue", j)
		RDB.Publish(CTX, "UpdateChannel", "+1")
		AlexeyFadeevNotification(request, ip)
		FlexLetter(request, ip)
	}else{
		return
	}
	if request.Trial{
		AddDelete(DeleteRequest{
			ID:         strconv.Itoa(droplet.ID),
			Expiration: time.Now().Add(time.Hour*24*3+2).Format(time.ANSIC),
		})
	}
}

func runCommand(ip, command string) error{
	key, err := ioutil.ReadFile("/root/.ssh/id_rsa")
	if err != nil {
		fmt.Println("RSA AUTH FAILED")
		return err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println("RSA AUTH FAILED")
		return err}
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		return err
	}
	defer client.Close()
	ss, err := client.NewSession()
	if err != nil {
		return err
	}
	defer ss.Close()
	var stdoutBuf bytes.Buffer
	ss.Stdout = &stdoutBuf
	err = ss.Run(command)
	return err
}

func floatingCheck(ip string)bool{
	con, err := net.Dial("tcp", ip+":7000")
	if err != nil {
		return false
	}
	fmt.Fprintf(con, "GET / HTTP/1.0\r\n\r\n")
	resp,_, err := bufio.NewReader(con).ReadLine()
	return bytes.Compare(resp, []byte{19, 0, 0, 0, 1, 4, 0, 0, 0, 0, 0, 0, 0, 2, 4, 0, 0, 0, 3, 0, 0, 0, 255})==0
}

func isactive(id string)bool{
	list, _, _ := DOClient.Droplets.List(CTX, &godo.ListOptions{PerPage: 1000})
	for _, v := range list{
		if strconv.Itoa(v.ID)==id{return true}
	}
	return false
}

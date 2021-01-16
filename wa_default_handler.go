package main

import (
	"fmt"
	"os"
	"log"
	"strings"

	whatsapp "github.com/Rhymen/go-whatsapp"
	req "github.com/imroc/req"

	"github.com/eFishery/NeMo/utils"
)


func currently_it_do_nothing(wac *whatsapp.Conn, RJID string) {
	phone_number := strings.Split(RJID, "@")[0]

	// if the user suddenly sent the image this will trigger error because there is no available session
	// need to test this
	_, err := loadSession(phone_number)
	if err != nil {
		log.Println("I don't know what you do but it do nothing")
		return
	}
}

func (wh *waHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	if !(message.Info.Timestamp < wh.startTime) {
		go currently_it_do_nothing(wh.c, message.Info.RemoteJid)
	}
}

func (wh *waHandler) HandleVideoMessage(message whatsapp.VideoMessage) {
	if !(message.Info.Timestamp < wh.startTime) {
		go currently_it_do_nothing(wh.c, message.Info.RemoteJid)
	}
}

func (wh *waHandler) HandleContactMessage(message whatsapp.ContactMessage) {
	if !(message.Info.Timestamp < wh.startTime) {
		go currently_it_do_nothing(wh.c, message.Info.RemoteJid)
	}
}

// need to test if the greeting is function well and return nothing after send message
func greeting(wac *whatsapp.Conn, RJID string, message string){
	for gIndex := range(BuildGreetings) {
		for pIndex := range(BuildGreetings[gIndex].ExpectedUsers) {
			if(strings.Split(RJID, "@")[1] == "g.us" && BuildGreetings[gIndex].ExpectedUsers[pIndex] == "any"){
				fmt.Println("The any default message is enabled, and only accepted by direct message")
				return
			}
			if(BuildGreetings[gIndex].ExpectedUsers[pIndex] == RJID || BuildGreetings[gIndex].ExpectedUsers[pIndex] == "any"){
				url := BuildGreetings[gIndex].Webhook.URL

				logGreeting := utils.LogGreeting {
					Message: message,
					PhoneNumber: strings.Split(RJID, "@")[0],
				}

				switch BuildGreetings[gIndex].Webhook.Service {
				case "DISCORD":
					_, LogErr := LogToDiscord(url, logGreeting)
					if LogErr != nil {
						log.Println("Fail to Log : " + LogErr.Error())
					}
				case "WEBHOOK":
					_, LogErr := LogToWebhook(url, logGreeting)
					if LogErr != nil {
						log.Println("Fail to Log : " + LogErr.Error())
					}
				}

				go sendMessage(wac, BuildGreetings[gIndex].Message, RJID)

				return
			}
		}
	}
}

func sendImage(wac *whatsapp.Conn, RJID string, imageUrl string, caption string) {
	// best way to stream image and send
	// don't have time to backup, so I just comment this haha
	// reqImg, err := http.Get(imageUrl)
	// if err != nil {
	//     log.Fatalf("http.Get -> %v", err)
	// }

	// reqImg.Body.Close()
	// img, err := ioutil.ReadAll(reqImg.Body)
	// if err != nil {
	//     log.Fatalf("ioutil.ReadAll -> %v", err)
	// }

	r, _ := req.Get(imageUrl)
	r.ToFile("/tmp/tmp.png")

	img, err := os.Open("/tmp/tmp.png")

	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	msg := whatsapp.ImageMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: RJID,
		},
		Type:    "image/jpeg",
		Caption: caption,
		// Content: bytes.NewReader(img),
		Content: img,
	}

	log.Println("sent the image")
	msgId, err := wac.Send(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgId)
	}
}
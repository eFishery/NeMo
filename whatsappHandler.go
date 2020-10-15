package main

import (
	"encoding/json"
	"io/ioutil"

	"os"
	"fmt"
	"log"
	"time"
	"strings"
	"regexp"
	"strconv"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

func (wh *waHandler)  HandleImageMessage(message whatsapp.ImageMessage) {
	if !(message.Info.Timestamp < wh.startTime) {

		phone_number := strings.Split(message.Info.RemoteJid, "@")[0]

		// if the user suddenly sent the image this will trigger error because there is no available session
		Sessions, err := loadSession(phone_number)
		if err != nil {
			log.Println(phone_number + " sent me image but it does nothing")
			return
		}

		if Sessions.CurrentProcess == "" {
			log.Println(phone_number + " sent me image but it does nothing")
			return
		}

		if Sessions.ProcessStatus == "WAIT_ANSWER" {

			var coral Coral
			coral.getCoral(Sessions.CurrentProcess)
			sIndex := Sessions.CurrentQuestionSlug

			// prevent for user put image on any rule
			if coral.Process.Questions[sIndex].Question.Validation.Rule != "image" {
				go sendMessage(wh.c, coral.Process.Questions[sIndex].Question.Validation.Message, message.Info.RemoteJid)
			}

			data, err := message.Download()
			if err != nil {
				if err != whatsapp.ErrMediaDownloadFailedWith410 && err != whatsapp.ErrMediaDownloadFailedWith404 {
					return
				}
				if _, err = wh.c.LoadMediaInfo(message.Info.RemoteJid, message.Info.Id, strconv.FormatBool(message.Info.FromMe)); err == nil {
					data, err = message.Download()
					if err != nil {
						return
					}
				}
			}
			filename := fmt.Sprintf("%v/%v.%v", os.TempDir(), message.Info.Id, strings.Split(message.Type, "/")[1])
			file, err := os.Create(filename)
			defer file.Close()
			if err != nil {
				return
			}
			_, err = file.Write(data)
			if err != nil {
				return
			}
			log.Printf("%v %v\n\timage received, saved at:%v\n", message.Info.Timestamp, message.Info.RemoteJid, filename)
			
			uploadS3 := AddFileToS3(filename)

			log.Println("Files Uploaded and here is the link : " + uploadS3)
			
			reply := "terminate"

			waktu, err := time.Parse(time.RFC3339, Sessions.Expired)

			if err != nil {
				fmt.Println(err)
			}

			if waktu.Before(time.Now()) {
				reply = "Sesi anda susah habis, silahkan ulangi lagi"
				Sessions.ProcessStatus = "DONE"
				file, _ := json.MarshalIndent(Sessions, "", " ")
				_ = ioutil.WriteFile(fileSession(phone_number), file, 0644)

				if reply != "timeout" {
					go sendMessage(wh.c, reply, message.Info.RemoteJid)
				}

				return
			}

			if sIndex >= (len(coral.Process.Questions)-1) {
				reply = coral.Process.EndMessage
				Sessions.ProcessStatus = "DONE"
				Sessions.Finished = time.Now().Format(time.RFC3339)
			}else{
				reply = coral.Process.Questions[sIndex+1].Question.Asking
				Sessions.ProcessStatus = "NEXT"
				Sessions.CurrentQuestionSlug = sIndex+1
			}

			dataBaru := Data{
				Slug: coral.Process.Questions[sIndex].Question.Slug,
				Question: coral.Process.Questions[sIndex].Question.Asking,
				Answer: uploadS3,
				Created: time.Now().Format(time.RFC3339),
			}

			Sessions.Datas = append(Sessions.Datas, dataBaru)

			go saveSession(Sessions, phone_number)

			if coral.Commands.Record {
				switch coral.Webhook.Service {
				case "DISCORD":
					_, errSent := SentToDiscord(coral.Webhook.URL, Sessions)
					if errSent != nil {
						log.Println(errSent.Error())
					}
				case "WEBHOOK":
					_, errSent := SentToWebhook(coral.Webhook.URL, Sessions)
					if errSent != nil {
						log.Println(errSent.Error())
					}
				}
			}

			if reply != "timeout" {
				if Sessions.ProcessStatus != "WAIT_ANSWER" {
					go sendMessage(wh.c, reply, message.Info.RemoteJid)

					if Sessions.ProcessStatus != "DONE" {
						log.Println(Sessions.ProcessStatus)
						Sessions.ProcessStatus = "WAIT_ANSWER"
					}

					go saveSession(Sessions, phone_number)
				}
			}
		}
	}
}

func (wh *waHandler) HandleTextMessage(message whatsapp.TextMessage) {

	var Sessions Session

	// Check the existing commands
	for index := range(BuildCommands) {

		// if the user force a new command while in the progress of session, break session and create a new one

		phone_number := strings.Split(message.Info.RemoteJid, "@")[0]

		cur_cmd := fmt.Sprintf("%s%s", BuildCommands[index].Prefix, BuildCommands[index].Command )
		if !strings.Contains(strings.ToLower(message.Text), cur_cmd) || message.Info.Timestamp < wh.startTime {
			continue
		}

		reply := "timeout"
		process := BuildCommands[index].RunProcess
		var coral Coral
		coral.getCoral(process)

		if len(coral.ExpectedUsers) > 0 {
			for usersIndex := range(coral.ExpectedUsers) {
				if coral.ExpectedUsers[usersIndex] == phone_number || coral.ExpectedUsers[usersIndex] == "any" {
					break
				}
				if len(coral.ExpectedUsers)-1 == usersIndex {
					log.Println(phone_number + " Trying to command " + cur_cmd + " for coral " + process + ", but not as expected users")
					return
				}
			}
		}

		reply, parserErr := nemoParser(BuildCommands[index].Message, Sessions)
		if parserErr != nil {
			log.Println(parserErr.Error())
			return
		}

		if reply != "timeout" {
			go sendMessage(wh.c, reply, message.Info.RemoteJid)
		}

		time.Sleep(time.Duration(3) * time.Second)

		if BuildCommands[index].RunProcess != "" && coral.Commands.RunProcess {
			savedSession := newSession(phone_number, process, coral.Process.Timeout)

			reply = coral.Process.Questions[savedSession.CurrentQuestionSlug].Question.Asking

			if reply != "timeout" {
				go sendMessage(wh.c, reply, message.Info.RemoteJid)
			}
		}

		return
	}

	// Check the message replied
	if !(message.Info.Timestamp < wh.startTime) {

		log.Println(message.Info.RemoteJid + ": " + message.Text)

		// check the previous message who send the message, if bot, check the message, if still same, just keep silent, if not continue
		// if user reply then can do
		
		phone_number := strings.Split(message.Info.RemoteJid, "@")[0]
		Sessions, err := loadSession(phone_number)
		if err != nil {
			go greeting(wh.c, message.Info.RemoteJid, message.Text)
			return
		}

		if Sessions.CurrentProcess == "" {
			go greeting(wh.c, message.Info.RemoteJid, message.Text)
			return
		}

		if Sessions.ProcessStatus == "DONE" || Sessions.ProcessStatus == "" {
			go greeting(wh.c, message.Info.RemoteJid, message.Text)
		}

		if Sessions.ProcessStatus == "WAIT_ANSWER" {
			reply := "terminate"
			sIndex := Sessions.CurrentQuestionSlug

			var coral Coral
			coral.getCoral(Sessions.CurrentProcess)

			waktu, err := time.Parse(time.RFC3339, Sessions.Expired)
			if err != nil {
				fmt.Println(err)
			}

			if waktu.Before(time.Now()) {
				Sessions.ProcessStatus = "DONE"
				Sessions.Finished = time.Now().Format(time.RFC3339)

				go saveSession(Sessions, phone_number)
				go sendMessage(wh.c, "Sesi anda susah habis, silahkan ulangi lagi", message.Info.RemoteJid)

				return
			}

			exit_cmd := fmt.Sprintf("%s%s", coral.Process.ExitCommand.Prefix, coral.Process.ExitCommand.Command)

			if message.Text == exit_cmd {
				Sessions.ProcessStatus = "DONE"
				Sessions.Finished = time.Now().Format(time.RFC3339)

				go saveSession(Sessions, phone_number)
				go sendMessage(wh.c, coral.Process.ExitCommand.Message, message.Info.RemoteJid)

				return
			}

			if coral.Process.Questions[sIndex].Question.Validation.Rule == "image" {
				go sendMessage(wh.c, coral.Process.Questions[sIndex].Question.Validation.Message, message.Info.RemoteJid)
				return
			}

			match, err := regexp.MatchString(coral.Process.Questions[sIndex].Question.Validation.Rule, message.Text)
			if !match {
				go sendMessage(wh.c, coral.Process.Questions[sIndex].Question.Validation.Message, message.Info.RemoteJid)
				return
			}

			if sIndex >= (len(coral.Process.Questions)-1) {
				reply = coral.Process.EndMessage
				Sessions.ProcessStatus = "DONE"
				Sessions.Finished = time.Now().Format(time.RFC3339)
			}else{
				reply = coral.Process.Questions[sIndex+1].Question.Asking
				Sessions.ProcessStatus = "NEXT"
				Sessions.CurrentQuestionSlug = sIndex+1
			}

			dataBaru := Data{
				Slug: coral.Process.Questions[sIndex].Question.Slug,
				Question: coral.Process.Questions[sIndex].Question.Asking,
				Answer: message.Text,
				Created: time.Now().Format(time.RFC3339),
			}

			Sessions.Datas = append(Sessions.Datas, dataBaru)

			go saveSession(Sessions, phone_number)

			if coral.Commands.Record {
				switch coral.Webhook.Service {
				case "DISCORD":
					_, errSent := SentToDiscord(coral.Webhook.URL, Sessions)
					if errSent != nil {
						log.Println(errSent.Error())
					}
				case "WEBHOOK":
					_, errSent := SentToWebhook(coral.Webhook.URL, Sessions)
					if errSent != nil {
						log.Println(errSent.Error())
					}
				}
			}

			if reply != "timeout" {
				if Sessions.ProcessStatus != "WAIT_ANSWER" {
					go sendMessage(wh.c, reply, message.Info.RemoteJid)

					if Sessions.ProcessStatus != "DONE" {
						Sessions.ProcessStatus = "WAIT_ANSWER"
					}

					file, _ := json.MarshalIndent(Sessions, "", " ")
					_ = ioutil.WriteFile(fileSession(phone_number), file, 0644)
				}
			}
		}
	}
	return
}

func currently_it_do_nothing(wac *whatsapp.Conn, RJID string) {
	phone_number := strings.Split(RJID, "@")[0]

	// if the user suddenly sent the image this will trigger error because there is no available session
	// need to test this
	_, err := loadSession(phone_number)
	if err != nil {
		go sendMessage(wac, "I don't know what you do but it do nothing", RJID)
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

				logGreeting := LogGreeting {
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
package main

import (

	"fmt"
	"log"
	"time"
	"strings"
	"regexp"

	whatsapp "github.com/Rhymen/go-whatsapp"

	"github.com/eFishery/NeMo/utils"
)


func (wh *waHandler) HandleTextMessage(message whatsapp.TextMessage) {

	var Sessions utils.Session

	// Check the existing commands
	for index := range BuildCommands {

		// if the user force a new command while in the progress of session, break session and create a new one

		phone_number := strings.Split(message.Info.RemoteJid, "@")[0]

		cur_cmd := fmt.Sprintf("%s%s", BuildCommands[index].Prefix, BuildCommands[index].Command )
		if !strings.Contains(strings.ToLower(message.Text), cur_cmd) || message.Info.Timestamp < wh.startTime {
			continue
		}

		reply := "timeout"
		process := BuildCommands[index].RunProcess
		var coral utils.Coral
		coral.GetCoral(process)

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

		sepparator := fmt.Sprintf("%s%s ", coral.Commands.Prefix, coral.Commands.Command)
		var question = ""
		if len(strings.Split(message.Text, strings.ToLower(sepparator))) > 1 {
			question = strings.Split(message.Text, strings.ToLower(sepparator))[1]
		}

		Sessions.PhoneNumber = phone_number
		Sessions.CurrentProcess = BuildCommands[index].RunProcess
		Sessions.Argument = question
		Sessions.ProcessStatus = "WAIT_ANSWER"
		reply, commonResponse, parserErr := commandParser(BuildCommands[index].Message, Sessions)
		if parserErr != nil {
			log.Println(parserErr.Error())
			return
		}

		if reply != "timeout" {
			for indexImage := range(commonResponse.Images) {
				go sendImage(wh.c, message.Info.RemoteJid, commonResponse.Images[indexImage].URL, commonResponse.Images[indexImage].Caption)
			}
			go sendMessage(wh.c, reply, message.Info.RemoteJid)			
		}

		time.Sleep(time.Duration(3) * time.Second)

		if BuildCommands[index].RunProcess != "" && coral.Commands.RunProcess {
			savedSession := newSession(Sessions, coral.Process.Timeout)

			reply = coral.Process.Questions[savedSession.CurrentQuestionSlug].Question.Asking

			replyS, _, parserErr := processParser(reply, Sessions)
			reply = replyS
			if parserErr != nil {
				log.Println("Error while parsing the end message of the process")
			}

			if reply != "timeout" {
				go sendMessage(wh.c, reply, message.Info.RemoteJid)

				defer saveSession(savedSession, phone_number)
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
			log.Println("return nothing greeting")
			return
		}

		if Sessions.CurrentProcess == "" {
			go greeting(wh.c, message.Info.RemoteJid, message.Text)
			log.Println("return nothing current process nothing")
			return
		}

		if Sessions.ProcessStatus == "DONE" || Sessions.ProcessStatus == "" {
			go greeting(wh.c, message.Info.RemoteJid, message.Text)
		}

		if Sessions.ProcessStatus == "WAIT_ANSWER" {
			reply := "terminate"
			sIndex := Sessions.CurrentQuestionSlug

			var coral utils.Coral
			coral.GetCoral(Sessions.CurrentProcess)

			waktu, err := time.Parse(time.RFC3339, Sessions.Expired)
			if err != nil {
				fmt.Println(err)
			}

			if waktu.Before(time.Now()) {
				Sessions.ProcessStatus = "DONE"
				Sessions.Finished = time.Now().Format(time.RFC3339)

				defer saveSession(Sessions, phone_number)
				go sendMessage(wh.c, "Sesi anda susah habis, silahkan ulangi lagi", message.Info.RemoteJid)

				return
			}

			exit_cmd := fmt.Sprintf("%s%s", coral.Process.ExitCommand.Prefix, coral.Process.ExitCommand.Command)

			if message.Text == exit_cmd {
				Sessions.ProcessStatus = "DONE"
				Sessions.Finished = time.Now().Format(time.RFC3339)

				defer saveSession(Sessions, phone_number)
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

			questionParsed, _, parserErr := processParser(coral.Process.Questions[sIndex].Question.Asking, Sessions)
			if parserErr != nil {
				log.Println("Error while parsing the end message of the process")
			}

			dataBaru := utils.Data{
				Slug: coral.Process.Questions[sIndex].Question.Slug,
				Question: questionParsed,
				Answer: message.Text,
				Created: time.Now().Format(time.RFC3339),
			}

			Sessions.Datas = append(Sessions.Datas, dataBaru)

			if coral.Process.Log {
				logged := SentTo(coral.Log.Service, coral.Log.URL, Sessions)
				if logged {
					log.Println("Data sucess sent to " + coral.Log.Service)
				}
			}

			if sIndex >= (len(coral.Process.Questions)-1) {
				if coral.Process.Record {
					webhook := SentTo(coral.Webhook.Service, coral.Webhook.URL, Sessions)
					if webhook {
						log.Println("Data sucess sent to " + coral.Webhook.Service)
					}
				}
			}

			if reply != "timeout" {
				if Sessions.ProcessStatus != "WAIT_ANSWER" {

					if Sessions.ProcessStatus != "DONE" {
						Sessions.ProcessStatus = "WAIT_ANSWER"
					}

					defer saveSession(Sessions, phone_number)

					replyS, _, parserErr := processParser(reply, Sessions)
					reply = replyS
					if parserErr != nil {
						log.Println("Error while parsing the end message of the process")
					}

					go sendMessage(wh.c, reply, message.Info.RemoteJid)
				}
			}
		}
	}
	return
}

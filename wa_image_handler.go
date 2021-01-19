package main

import (
	"encoding/json"
	"io/ioutil"

	"os"
	"fmt"
	"log"
	"time"
	"strings"
	"strconv"

	whatsapp "github.com/Rhymen/go-whatsapp"

	"github.com/eFishery/NeMo/utils"
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

			var coral utils.Coral
			coral.GetCoral(Sessions.CurrentProcess)
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
			
			uploadS3 := Settings.AddFileToS3(filename)

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
				_ = ioutil.WriteFile(utils.FileSession(phone_number), file, 0644)

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

			dataBaru := utils.Data{
				Slug: coral.Process.Questions[sIndex].Question.Slug,
				Question: coral.Process.Questions[sIndex].Question.Asking,
				Answer: uploadS3,
				Created: time.Now().Format(time.RFC3339),
			}

			Sessions.Datas = append(Sessions.Datas, dataBaru)

			go saveSession(Sessions, phone_number)

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
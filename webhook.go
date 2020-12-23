package main

import (
	"log"
	req "github.com/imroc/req"

	"github.com/eFishery/NeMo/utils"
)

func SentToWebhook(url string, Sessions utils.Session) (int, error) {
	r, err := req.Post(url, req.BodyJSON(Sessions))
	if err != nil {
		return 500, err
	}

	resp := r.Response()

	return resp.StatusCode, nil
}

func SentToDiscord(url string, Sessions utils.Session) (bool, error) {
	// req.Debug = true
	compiled_message := "\n[" + Sessions.Created + "]\n" + Sessions.CurrentProcess
	for index := range(Sessions.Datas) {
		compiled_message = compiled_message + "\n[" + Sessions.Datas[index].Created + "]"
		compiled_message = compiled_message + "\nnemo: " + Sessions.Datas[index].Slug + " - " + Sessions.Datas[index].Question
		compiled_message = compiled_message + "\n" + Sessions.PhoneNumber + ": " + Sessions.Datas[index].Answer
	}

	var Discord = utils.Discord {
		Content: compiled_message,
	}
	_, err := req.Post(url, req.BodyJSON(Discord))

	if err != nil {
		return false, err
	}

	return true, nil
}

func LogToWebhook(url string, logGreeting utils.LogGreeting) (int, error) {
	r, err := req.Post(url, req.BodyJSON(logGreeting))
	if err != nil {
		return 500, err
	}

	resp := r.Response()

	return resp.StatusCode, nil
}

func LogToDiscord(url string, logGreeting utils.LogGreeting) (bool, error) {
	// req.Debug = true
	compiled_message := logGreeting.PhoneNumber + " replied with " + logGreeting.Message

	var Discord = utils.Discord {
		Content: compiled_message,
	}
	_, err := req.Post(url, req.BodyJSON(Discord))

	if err != nil {
		log.Println("ERROR: "+ err.Error())
		return false, err
	}

	return true, nil
}

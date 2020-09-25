package main

import (
	"encoding/json"
	"io/ioutil"

	"log"
	"fmt"
	"time"
)

func newSession(phone_number string, current_process string, timeout int) Session{

	savedSession := Session {
		PhoneNumber: phone_number,
		CurrentProcess: current_process,
		CurrentQuestionSlug: 0,
		ProcessStatus: "WAIT_ANSWER",
		Created: time.Now().Format(time.RFC3339),
		Expired: time.Now().Add(time.Second * time.Duration(timeout)).Format(time.RFC3339),
	}

	file, _ := json.MarshalIndent(savedSession, "", " ")
	_ = ioutil.WriteFile(fileSession(phone_number), file, 0644)

	return savedSession
}

func loadSession(phone_number string) (Session, error) {
	var s Session
	file_session, err := ioutil.ReadFile(fileSession(phone_number))

	if err != nil {
		log.Println("Create a new file")
		file, _ := json.MarshalIndent(Session{}, "", " ")
		_ = ioutil.WriteFile(fileSession(phone_number), file, 0644)
		return s, fmt.Errorf("Session hasn't been created")	
	}

	jsonErr := json.Unmarshal(file_session, &s)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return s, nil
}

func saveSession(s Session, phone_number string) {
	file, _ := json.MarshalIndent(s, "", " ")
	_ = ioutil.WriteFile(fileSession(phone_number), file, 0644)	
}
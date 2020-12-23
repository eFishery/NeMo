package main

import (
	"encoding/gob"
	"math/rand"
	"os/signal"

	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"syscall"

	whatsapp "github.com/Rhymen/go-whatsapp"
	cron "github.com/robfig/cron/v3"
	godotenv "github.com/joho/godotenv"
	// "github.com/davecgh/go-spew/spew"

	"github.com/eFishery/NeMo/utils"
)

type waHandler struct {
	c         *whatsapp.Conn
	startTime uint64
	chats map[string]struct{}

}

func init() {
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

var Settings *utils.Setting
var BuildGreetings []utils.BuildGreeting
var BuildCommands []utils.BuildCommand

func main() {

	if len(os.Args) < 2 {
		log.Fatal("You need to specify the phone number/ file session")
	}

	sender := os.Args[1]

	Settings = utils.LoadSetting()
	Settings.Builder()
	BuildGreetings = utils.ReadGreetingsFile()

	jadwal := cron.New()

	BuildCommands = utils.ReadBuildCommandsFiles()

	wac, err := whatsapp.NewConn(5 * time.Second)
	// wac.SetClientVersion(0, 4, 2080)
	wac.SetClientVersion(2, 2035, 14)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	handler := &waHandler{wac, uint64(time.Now().Unix()), make(map[string]struct{})}
	wac.AddHandler(handler)

	if err := login(wac, sender); err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}
	
	isLoaded, Schedules := utils.ReadScheduleFiles()
	if !isLoaded {
		log.Println("Can't read Schedule Files")
		return
	}

	for index := range(Schedules) {
		phone_numbers := Schedules[index].ExpectedUsers
		process_name := Schedules[index].ProcessName
		log.Println("Read the schedule to run with cron formula " + Schedules[index].Rule)
		jadwal.AddFunc(Schedules[index].Rule, func(){
			log.Println("Run Schedule " + process_name)
			for pIndex := range(phone_numbers) {
				go sendMessage(wac, Schedules[index].Message, phone_numbers[pIndex])
			}
		})
	}

	jadwal.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down now.")
	jadwal.Stop()
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session, sender); err != nil {
		log.Fatalf("error saving session: %v", err)
	}

	os.Exit(1)
}

func sendMessage(wac *whatsapp.Conn, message string, RJID string) {

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: RJID,
		},
		Text: message,
	}

	// prevent disorder post while you need the chat in order
	min, max, limit_max := Settings.RandMin, Settings.RandMax, Settings.LimitRandMax

	kata := strings.Split(message, " ")

	// definition : kata = 50 and limit_max = 5
	// if min is 2 then 2 + (50/2) = 27
	// if max is 4 then 4 + (50/2) = 29
	// if 29 > 5 then max is 5
	min = min + (len(kata)/2)
	max = max + (len(kata)/2)
	if max > limit_max {
		min = limit_max - 2
		max = limit_max
	}
	waitSec := rand.Intn(max-min)+min
	log.Printf("Randomly paused %d for throtling", waitSec)
	wac.Presence(RJID, whatsapp.PresenceComposing)
	time.Sleep(time.Duration(waitSec) * time.Second)

	msgId, err := wac.Send(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgId)
	}
}

func login(wac *whatsapp.Conn, phone_number string) error {
	session, err := readSession(phone_number)
	if err == nil {
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		log.Println("Session is not available, please prepare the session file")
		os.Exit(1)
	}

	err = writeSession(session, phone_number)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession(phone_number string) (whatsapp.Session, error) {
	session := whatsapp.Session{}
	log.Println("Trying to get the session " + getSessionName(phone_number))
	file, err := os.Open(getSessionName(phone_number))
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session, phone_number string) error {
	file, err := os.Create(getSessionName(phone_number))
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

func getSessionName(phone_number string) string {
	if _, err := os.Stat(Settings.SessionsDir); os.IsNotExist(err) {
		os.MkdirAll(Settings.SessionsDir, os.ModePerm)
	}
	return Settings.SessionsDir + "/" + phone_number + ".gob"
}

func exit(wac *whatsapp.Conn, phone_number string) {
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session, phone_number); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
	os.Exit(1)
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}
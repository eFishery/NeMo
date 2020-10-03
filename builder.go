package main

import (
	"io/ioutil"
	"os"
	"log"
	"fmt"
	"regexp"
	"encoding/json"
)

// Test if filename is a valid yaml file (ends with `yml` or `yaml`).
// If it's matching then returns two strings: filename and extension
// If it's not matching these values are empty
func matchYAMLFile(filename string) bool {
  re := regexp.MustCompile(`^.*\.(ya?ml)$`)
  match := re.MatchString(filename)
  return match
}


func builder() {
	linter := builder_linter_all()
	if len(linter) > 0 {
		for i:= range(linter){
			log.Println(linter[i])
		}
		os.Exit(1)
	}

	files,_ := ioutil.ReadDir(Settings.CoralDir)
	var dataSetCommands []BuildCommand
	var dataSetSchedules []Schedule
	var dataSetGreetings []BuildGreeting
	saveToFile := true
	for _, file := range files {
		saveToCommand := true
		saveToSchedule := true
		if matchYAMLFile(file.Name()) {
			log.Println("Build file " + file.Name())
			processName := file.Name()
			var coral Coral
			coral.getCoral(processName)

			var commandCompile = BuildCommand {
				Prefix: coral.Commands.Prefix,
				Command: coral.Commands.Command,
				Record: coral.Commands.Record,
				RunProcess: processName,
				Message: coral.Commands.Message,
			}

			if coral.valCommands() {
				for index := range(dataSetCommands) {
					if dataSetCommands[index].Prefix == commandCompile.Prefix && dataSetCommands[index].Command == commandCompile.Command {
						log.Println("Command " + commandCompile.Prefix + commandCompile.Command + " is skipped because already exist in process " + dataSetCommands[index].RunProcess)
						saveToCommand = false
						break
					}
				}
				if saveToCommand {
					dataSetCommands = append(dataSetCommands, commandCompile)
				}
			}

			if coral.valSchedule() {

				var ExpectedUsers []string
				for ii := range(coral.ExpectedUsers) {
					ExpectedUsers = append(ExpectedUsers, fmt.Sprintf("%s@s.whatsapp.net", coral.ExpectedUsers[ii]))
				}
				var scheduleCompile = Schedule {
					Rule: coral.Schedule.Rule,
					ProcessName: processName,
					Message: coral.Schedule.Message,
					ExpectedUsers: ExpectedUsers,
					Sender: coral.Schedule.Sender,
				}
				for index := range(dataSetSchedules) {
					if dataSetSchedules[index].Rule == scheduleCompile.Rule {
						log.Println("Schedule " + scheduleCompile.Rule + " is skipped because already exist in process " + dataSetSchedules[index].ProcessName)
						saveToSchedule = false
						break
					}
				}
				if saveToSchedule {
					dataSetSchedules = append(dataSetSchedules, scheduleCompile)
				}
			}

			if coral.DefaultGreeting.Message != "" {

				var ExpectedUsers []string
				for ii := range(coral.ExpectedUsers) {
					ExpectedUsers = append(ExpectedUsers, fmt.Sprintf("%s@s.whatsapp.net", coral.ExpectedUsers[ii]))
				}
				var defaultGreetingCompile = BuildGreeting {
					ProcessName: processName,
					Message: coral.DefaultGreeting.Message,
					Webhook: coral.DefaultGreeting.Webhook,
					ExpectedUsers: ExpectedUsers,
				}
				dataSetGreetings = append(dataSetGreetings, defaultGreetingCompile)
			}

		}else{
			log.Println("Skip file " + file.Name() + " is not ended with .yml nor .yaml")
		}
	}

	if saveToFile {
		commands, _ := json.MarshalIndent(dataSetCommands, "", " ")
		_ = ioutil.WriteFile(Settings.BuildDir + "/commands.json", commands, 0644)

		schedules, _ := json.MarshalIndent(dataSetSchedules, "", " ")
		_ = ioutil.WriteFile(Settings.BuildDir + "/schedules.json", schedules, 0644)

		greetings, _ := json.MarshalIndent(dataSetGreetings, "", " ")
		_ = ioutil.WriteFile(Settings.BuildDir + "/greetings.json", greetings, 0644)
	}
}

func builder_linter_all() []string {
	var result []string
	files,_ := ioutil.ReadDir(Settings.CoralDir)
	for _, file := range files {
		if matchYAMLFile(file.Name()) {
			var coral Coral

			coral.getCoral(file.Name())

			if !coral.valAuthor() {
				result = append(result, file.Name() + ": Author must complete")
			}

			if coral.CommandExist() {
				if coral.Commands.Prefix == "" || coral.Commands.Command == "" {
					result = append(result, file.Name() + ": Commands must have prefix and command")
				}
			}

			if coral.Process.ExitCommand.Command != "" || coral.Process.ExitCommand.Prefix != "" || coral.Process.ExitCommand.Message != ""{
				if !coral.Commands.RunProcess {
					result = append(result, file.Name() + ": The Process command is exist, but the run process is false")
				}
			}

			if coral.Commands.RunProcess {
				if coral.Process.ExitCommand.Command == "" || coral.Process.ExitCommand.Prefix == "" || coral.Process.ExitCommand.Message == ""{
					result = append(result, file.Name() + ": You must set exit command value in process")
				}

				if coral.Process.Timeout == 0 {
					result = append(result, file.Name() + ": You must set a value for timeout second in process")
				}

				if len(coral.Process.Questions) == 0 {
					result = append(result, file.Name() + ": You need to have a question in process")
				}

				if coral.Process.EndMessage == "" {
					result = append(result, file.Name() + ": You need to set a value for End Message in process")
				}
			}

			if coral.Commands.Record {
				if coral.Webhook.Service == "" || coral.Webhook.URL == "" {
					result = append(result, file.Name() + ": You need to set a value Webhook")
				}
			}

		}else{
			log.Println("Skip file " + file.Name() + " is not ended with .yml nor .yaml")
		}
	}
	return result
}

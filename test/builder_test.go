package test

import (
	"io/ioutil"
	"testing"
	"os"

	"github.com/eFishery/NeMo/utils"
)

func TestBuildCommandsAutoCreated( t *testing.T) {
	Settings = LoadSetting()

	_ = ioutil.WriteFile(Settings.BuildDir + "/commands.json", []byte(""), 0644)

	fileCommands := Settings.BuildDir + "/commands.json"

	e := os.Remove(fileCommands)
	if e != nil {
		t.Errorf("Can not Delete file, please make sure file commands.json in " + Settings.BuildDir + " is exist in order to run this test")
		os.Exit(1)
    }

	if _, err := os.Stat(fileCommands); os.IsNotExist(err) {
		builder()

		if _, err := os.Stat(fileCommands); os.IsNotExist(err) {
			t.Errorf("The file is not auto created, the file commands.json suppose to created")
		}else{
			t.Logf("Testing Success")
		}
	}else{
		t.Errorf("The file is suppose to be deleted, but just by pass")
	}
}

func TestSchedulesCommandsAutoCreated( t *testing.T) {
	Settings = LoadSetting()

	_ = ioutil.WriteFile(Settings.BuildDir + "/schedules.json", []byte(""), 0644)

	fileCommands := Settings.BuildDir + "/schedules.json"

	e := os.Remove(fileCommands)
	if e != nil {
		t.Errorf("Can not Delete file, please make sure file schedules.json in " + Settings.BuildDir + " is exist in order to run this test")
		os.Exit(1)
    }

	if _, err := os.Stat(fileCommands); os.IsNotExist(err) {
		builder()

		if _, err := os.Stat(fileCommands); os.IsNotExist(err) {
			t.Errorf("The file is not auto created, the file schedules.json suppose to created")
		}else{
			t.Logf("Testing Success")
		}
	}else{
		t.Errorf("The file is suppose to be deleted, but just by pass")
	}
}

func TestScheduleDuplicate( t *testing.T) {
	Settings = LoadSetting()
	linter := builder_linter_all()

	if len(linter) > 0 {
		for i:= range(linter){
			t.Errorf(linter[i])
		}
		
	}
}
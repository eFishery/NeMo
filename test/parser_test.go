package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/eFishery/NeMo/utils"
)

const (
	envTestURL1 = "TEST_URL_1"
	envTestURL2 = "TEST_URL_2"
	envTestURL3 = "TEST_URL_3"
)

var (
	testURL1, testURL2, testURL3 string
)

func TestNemoParser(t *testing.T) {
	setupTestNemoParser(t)
	tests := []struct {
		name   string
		pesan  string
		res    string
		config string
	}{
		{
			name:  "one_url",
			pesan: fmt.Sprintf("hello {{%s}}", testURL1),
			res:   "hello 1",
		},
		{
			name:  "two_url",
			pesan: fmt.Sprintf("hello {{%s}} hello {{%s}}", testURL1, testURL2),
			res:   fmt.Sprintf("hello 1 hello %s", errRespNotSupported(testURL2)),
		},
		{
			name:   "three_url_from_file",
			pesan:  fmt.Sprintf("hello {{%s}} hello {{%s}} hello {{%s}}", testURL1, testURL2, testURL3),
			res:    "hello 1 hello 2 hello 3",
			config: "test/config/keys-example.json",
		},
	}
	sess := Session{
		PhoneNumber:         "628123123123",
		CurrentProcess:      "basic",
		CurrentQuestionSlug: 1,
		ProcessStatus:       "DONE",
		Datas:               []Data{},
		Sent:                "",
		SentTo:              "",
		Created:             "2020-09-25T13:54:19+07:00",
		Expired:             "2020-09-25T13:59:19+07:00",
		Finished:            "2020-09-25T13:55:17+07:00",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(defSupportedRespKeysConfig, tt.config)
			pesan, err := nemoParser(tt.pesan, sess)
			if err != nil {
				t.Error(err)
				return
			}
			if pesan != tt.res {
				t.Errorf("\nwanted:\n%s\ngot:\n%s", tt.res, pesan)
			}
		})
	}
}

// sets up required value here
// TODO unset env?
func setupTestNemoParser(t *testing.T) {
	testURL1 = os.Getenv(envTestURL1)
	if testURL1 == "" {
		t.Fatalf("%s is not set", envTestURL1)
	}
	testURL2 = os.Getenv(envTestURL2)
	if testURL2 == "" {
		t.Fatalf("%s is not set", envTestURL2)
	}
	testURL3 = os.Getenv(envTestURL3)
	if testURL3 == "" {
		t.Fatalf("%s is not set", envTestURL3)
	}
}

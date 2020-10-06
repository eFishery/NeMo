package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"reflect"
	"testing"

	wa "github.com/Rhymen/go-whatsapp"
)

const (
	testBaseURL = "http://127.0.0.1:6969"
)

var (
	testTextURL1  = testBaseURL + "/text/1"
	testTextURL2  = testBaseURL + "/text/2"
	testImageURL1 = testBaseURL + "/image/1"
)

func TestNemoParser(t *testing.T) {
	tests := []struct {
		name   string
		pesan  string
		txt    *wa.TextMessage
		imgs   map[int]ImageMessage
		config string
	}{
		{
			name:  "no_url",
			pesan: "hello",
			txt: &wa.TextMessage{
				Text: "hello",
			},
		},
		{
			name:  "text",
			pesan: fmt.Sprintf("hello {{%s}}", testTextURL1),
			txt: &wa.TextMessage{
				Text: "hello 1",
			},
		},
		{
			name:  "text_text",
			pesan: fmt.Sprintf("hello {{%s}} hello {{%s}}", testTextURL1, testTextURL2),
			txt: &wa.TextMessage{
				Text: fmt.Sprintf("hello 1 hello %s", errRespNotSupported(testTextURL2)),
			},
		},
		{
			name:   "format_config/text_text",
			pesan:  fmt.Sprintf("hello {{%s}} hello {{%s}}", testTextURL1, testTextURL2),
			config: "config/keys-example.json",
			txt: &wa.TextMessage{
				Text: "hello 1 hello 2",
			},
		},
		{
			name:  "image",
			pesan: fmt.Sprintf("{{%s}}", testImageURL1),
			imgs: map[int]ImageMessage{
				0: ImageMessage{
					Type:    "image/jpeg",
					Content: defaultImageBuffer(),
				},
			},
		},
		{
			name:  "image_image",
			pesan: fmt.Sprintf("hello {{%s}} image {{%s}}", testImageURL1, testImageURL1),
			imgs: map[int]ImageMessage{
				0: ImageMessage{
					Type:    "image/jpeg",
					Content: defaultImageBuffer(),
				},
				1: ImageMessage{
					Caption: "hello  image ",
					Type:    "image/jpeg",
					Content: defaultImageBuffer(),
				},
			},
		},
		{
			name:  "text_image",
			pesan: fmt.Sprintf("hello {{%s}} image {{%s}}", testTextURL1, testImageURL1),
			imgs: map[int]ImageMessage{
				0: ImageMessage{
					Caption: "hello 1 image ",
					Type:    "image/jpeg",
					Content: defaultImageBuffer(),
				},
			},
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
			os.Setenv(defSupportedRespKeysConfig, "")
			os.Setenv(defSupportedRespKeysConfig, tt.config)
			txt, imgs, err := nemoParser(tt.pesan, sess)
			if err != nil {
				t.Error(err)
				return
			}
			if txt != nil {
				if tt.txt == nil {
					t.Errorf("wanted no text, got %s text", txt.Text)
					return
				}
				if txt.Text != tt.txt.Text {
					t.Errorf("\nwanted:\n%s\ngot:\n%s", tt.txt.Text, txt.Text)
				}
			}
			if imgs != nil {
				if tt.imgs == nil {
					t.Errorf("wanted no image, got %d image", len(imgs))
				}
				if tt.imgs != nil && !reflect.DeepEqual(imgs, tt.imgs) {
					t.Error("images don't match")
				}
				return
			}
		})
	}
}

func defaultImageBuffer() []byte {
	m := image.NewRGBA(image.Rect(0, 0, 240, 240))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
	buffer := new(bytes.Buffer)
	jpeg.Encode(buffer, m, nil)
	return buffer.Bytes()
}

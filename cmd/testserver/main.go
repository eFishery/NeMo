package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"net/http"
	"strconv"
)

var (
	testTextURL1  = "/text/1"
	testTextURL2  = "/text/2"
	testTextURL3  = "/text/3"
	testImageURL1 = "/image/1"
)

// sets up required value here
func main() {
	http.HandleFunc(testTextURL1, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(json.RawMessage(`{"message": "1"}`))
	})
	http.HandleFunc(testTextURL2, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(json.RawMessage(`{"massage": "2"}`))
	})
	http.HandleFunc(testImageURL1, func(w http.ResponseWriter, r *http.Request) {
		m := image.NewRGBA(image.Rect(0, 0, 240, 240))
		blue := color.RGBA{0, 0, 255, 255}
		draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
		buffer := new(bytes.Buffer)
		err := jpeg.Encode(buffer, m, nil)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
		w.Write(buffer.Bytes())
	})

	http.ListenAndServe(":6969", nil)
}

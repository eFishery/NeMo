package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	wa "github.com/Rhymen/go-whatsapp"
	"github.com/imroc/req"
)

var (
	// TODO use struct for key priority to allow setting config via file
	// that might need extra config validation though
	defSupportedRespKeys = map[string]int{
		"message": 1,
	}
	defSupportedRespKeysConfig = "SUPPORTED_KEYS_CONFIG"

	// TODO handle error differently
	// these are only for visualization purpose
	errReqErr           = func(dst string) string { return fmt.Sprintf("[req error: %s]", dst) }
	errRespNotSupported = func(dst string) string { return fmt.Sprintf("[resp not supported error: %s]", dst) }
)

type ImageMessage struct {
	Caption   string
	Thumbnail []byte // set to nil, add on your own as I have no way to test this
	Type      string
	Content   []byte
}

// nemoParser parses pesan for URL in {{url}} format
// if no URL is found in pesan, pesan is returned as is
// if URL is found, try POST request to url
// if response contains supported JSON, pesan will be replaced with response
// currently only supports JSON response with message key
// if response contains image(s), image with caption will be sent instead of text
// if there are multiple URL that return images, it will send all images, only captioning the last image
// currently if duplicate URLs exist, the request will be repeated
func nemoParser(pesan string, Sessions Session) (*wa.TextMessage, map[int]ImageMessage, error) {
	var countImg int
	txt := &wa.TextMessage{}
	mapImgs := make(map[int]ImageMessage)
	urlCount := strings.Count(pesan, "{{")
	if urlCount == 0 {
		txt.Text = pesan
		return txt, nil, nil
	}
	for i := 0; i < urlCount; i++ {
		url := between(pesan, "{{", "}}")
		resp, err := req.Post(url, req.BodyJSON(Sessions))
		if err != nil {
			pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errReqErr(url), 1)
		}

		// check if json/image/else
		switch ct := resp.Response().Header.Get("Content-Type"); ct {
		case "application/json":
			var m map[string]interface{}
			resp.ToJSON(&m)

			// TODO maybe calls this in main to setup
			sk := supportKey(os.Getenv(defSupportedRespKeysConfig))
			k := lookupKey(m, sk)
			if k == "" {
				pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errRespNotSupported(url), 1)
				txt.Text = pesan
				continue
			}
			if m[k] != "" {
				pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), fmt.Sprintf("%v", m[k]), 1)
			}
			// if there is image, simply remove URL
			if len(mapImgs) > 0 {
				continue
			}
			txt.Text = pesan
		default:
			if strings.Contains(ct, "image") {
				b, err := resp.ToBytes()
				if err != nil {
					pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errReqErr(url), 1)
					continue
				}
				pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), "", 1)
				mapImgs[countImg] = ImageMessage{
					Content: b,
					Type:    ct,
				}
				countImg++
				continue
			}
			pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errRespNotSupported(url), 1)
		}
	}
	if len(mapImgs) > 0 {
		lastImg := mapImgs[countImg-1]
		lastImg.Caption = pesan
		mapImgs[countImg-1] = lastImg
		return nil, mapImgs, nil
	}

	return txt, nil, nil
}

// lookupKey looks up if key exists in a map based on priority
// return empty string if key is not supported
// currently only supports first level key
// might want to improve the algorithm
// i.e return once a supported key is found with the assumption
// that response body has commonly used keys to denote response message
// it also does not care if the config sets multiple keys with the same
// priority, it uses whichever is found first
func lookupKey(m map[string]interface{}, sup map[string]int) string {
	var key string
	var prio int
	for k := range m {
		for v, p := range sup {
			if k != v {
				continue
			}
			if key == "" {
				key = v
				prio = p
				continue
			} else if p < prio {
				key = v
				prio = p
				continue
			}
		}
	}
	return key
}

// supportKey sets up keys support from JSON config file
// TODO validate config file
func supportKey(file string) map[string]int {
	kp := make(map[string]int)
	if file == "" {
		return defSupportedRespKeys
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return defSupportedRespKeys
	}
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	for k, v := range m {
		pp, ok := v.(float64)
		if !ok {
			continue
		}
		kp[k] = int(pp)
	}
	return kp
}

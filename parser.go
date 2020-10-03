package main

import (
	"encoding/json"
	"fmt"
	req "github.com/imroc/req"
	"io/ioutil"
	"os"
	"strings"
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

// nemoParser parses pesan for URL in {{url}} format
// if no URL is found in pesan, pesan is returned as is
// if URL is found, try POST request to url
// currently only supports JSON response with message key
func nemoParser(pesan string, Sessions Session) (string, error) {
	urlCount := strings.Count(pesan, "{{")
	if urlCount == 0 {
		return pesan, nil
	}
	for i := 0; i < urlCount; i++ {
		url := between(pesan, "{{", "}}")
		r, err := req.Post(url, req.BodyJSON(Sessions))
		if err != nil {
			pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errReqErr(url), -1)
		}

		var m map[string]interface{}
		r.ToJSON(&m)

		// TODO maybe calls this in main to setup
		sk := supportKey(os.Getenv(defSupportedRespKeysConfig))
		k := lookupKey(m, sk)
		if k == "" {
			pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errRespNotSupported(url), -1)
			continue
		}
		if m[k] != "" {
			pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), fmt.Sprintf("%v", m[k]), -1)
		}
	}

	return pesan, nil
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
	for k, _ := range m {
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

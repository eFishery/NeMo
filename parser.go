package main

import (
	"encoding/json"
	"fmt"
	req "github.com/imroc/req"
	"io/ioutil"
	"strings"

	"github.com/eFishery/NeMo/utils"
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
func nemoParser(pesan string, Sessions utils.Session) (utils.CommonResponse, error) {
	var commonResponse utils.CommonResponse
	url := utils.Between(pesan, "{{", "}}")
	response, err := req.Post(url, req.BodyJSON(Sessions))
	if err != nil {
		pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errReqErr(url), -1)
	}

	response.ToJSON(&commonResponse)

	return commonResponse, nil
}

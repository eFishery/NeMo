package main

import (
	"fmt"
	req "github.com/imroc/req"
	"strings"
)

// TODO handle error differently
// these are only for visualization purpose
var (
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

		var pF pesanFetch
		r.ToJSON(&pF)

		if pF.Message != "" {
			pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), pF.Message, -1)
		} else {
			pesan = strings.Replace(pesan, fmt.Sprintf("{{%s}}", url), errRespNotSupported(url), -1)
		}
	}

	return pesan, nil
}

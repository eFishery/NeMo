package main

import (
	"fmt"
	req "github.com/imroc/req"
	"strings"
	"strconv"

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
func commandParser(reply string, Sessions utils.Session) (string, utils.CommonResponse, error) {
	var commonResponse utils.CommonResponse
	var coral utils.Coral
	coral.GetCoral(Sessions.CurrentProcess)

	parameter := strings.Split(Sessions.Argument, " ")
	formatCount := strings.Count(reply, "{{")
	if formatCount == 0 {
		return reply, commonResponse, nil
	}
	for i := 0; i < formatCount; i++ {
		value := utils.Between(reply, "{{", "}}")
		httpCount := strings.Count(value, "http")
		if httpCount > 0 {
			r, err := req.Post(value, req.BodyJSON(Sessions))
			if err != nil {
				reply = strings.Replace(reply, fmt.Sprintf("{{%s}}", value), errReqErr(value), -1)
			}

			r.ToJSON(&commonResponse)

			reply = strings.Replace(reply, fmt.Sprintf("{{%s}}", value), commonResponse.Message, -1)
		}else{
			if len(parameter) > 1 {
				sumFunction := strings.Count(reply, "sum(")
						
				if sumFunction > 0 {
					for i := 0; i < sumFunction; i++ {
						sumIndex := utils.Between(reply, "sum(", ")")
						clearsumIndex := strings.Replace(sumIndex, " ", "", -1)
						indexing := strings.Split(clearsumIndex, ",")
						calc := 0
						for iindexing := range indexing{
							parsingIndex, _ := strconv.Atoi(indexing[iindexing])
							calcValue, _ := strconv.Atoi(parameter[parsingIndex-1])
							calc = calc + calcValue
						}
						reply = strings.Replace(reply, fmt.Sprintf("{{sum(%s)}}", sumIndex), strconv.Itoa(calc), -1)
					}
				}else{
					iindex, _ := strconv.Atoi(value)
					replaceValue := parameter[iindex-1]
					reply = strings.Replace(reply, fmt.Sprintf("{{%s}}", value), replaceValue, -1)
				}
			}
		}
	}

	return reply, commonResponse, nil
}

func processParser(reply string, Sessions utils.Session) (string, utils.CommonResponse, error) {
	var commonResponse utils.CommonResponse
	var coral utils.Coral
	coral.GetCoral(Sessions.CurrentProcess)
	formatCount := strings.Count(reply, "{{")
	if formatCount > 0 {
		for i := 0; i < formatCount; i++ {
			slug := utils.Between(reply, "{{", "}}")
			httpCount := strings.Count(slug, "http")

			// fmt.Println("Trying to parse the ", slug)
			
			if httpCount > 0 {
				r, err := req.Post(slug, req.BodyJSON(Sessions))
				if err != nil {
					reply = strings.Replace(reply, fmt.Sprintf("{{%s}}", slug), errReqErr(slug), -1)
				}
	
				r.ToJSON(&commonResponse)
	
				reply = strings.Replace(reply, fmt.Sprintf("{{%s}}", slug), commonResponse.Message, -1)

				continue

			}else{
				sumFunction := strings.Count(reply, "sum(")

				slugIsNumber := false 
				if _, err := strconv.Atoi(slug); err == nil {
					parameter := strings.Split(Sessions.Argument, " ")
					argIndex, _ := strconv.Atoi(slug)
					replaceValue := parameter[argIndex-1]
					reply = strings.Replace(reply, fmt.Sprintf("{{%s}}", slug), replaceValue, -1)
					slugIsNumber = true

					continue
				}

				if sumFunction > 0 {
					sumSlug := utils.Between(reply, "sum(", ")")
					clearsumSlug := strings.Replace(sumSlug, " ", "", -1)
					slugs := strings.Split(clearsumSlug, ",")
					calc := 0
					for iSlug := range slugs{
						for slugIndexs := range coral.Process.Questions {
							if coral.Process.Questions[slugIndexs].Question.Slug == slugs[iSlug] {
								number, _ := strconv.Atoi(Sessions.Datas[slugIndexs].Answer)
								calc = calc + number
							}
						}						
					}

					reply = strings.Replace(reply, fmt.Sprintf("{{sum(%s)}}", sumSlug), strconv.Itoa(calc), -1)

					continue
				}

				if !slugIsNumber {
					answer := "Can't Found"
					for slugIndex := range coral.Process.Questions {						
						if coral.Process.Questions[slugIndex].Question.Slug == slug {
							answer = Sessions.Datas[slugIndex].Answer
						}
					}
					reply = strings.Replace(reply, fmt.Sprintf("{{%s}}", slug), answer, -1)

					continue
				}
			}
		}
	}

	return reply, commonResponse, nil
}


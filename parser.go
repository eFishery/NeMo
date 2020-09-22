package main

import (
	"strings"
	"log"
	req "github.com/imroc/req"
)

func nemoParser(pesan string, Sessions Session) string{
	for indexURL := 0; indexURL < strings.Count(pesan, "{{"); indexURL++ {
		url := between(pesan, "{{", "}}")

		r, err := req.Post(url, req.BodyJSON(Sessions))
		if err != nil {
			log.Fatal(err)
			return "timeout"
		}

		var pF pesanFetch

		r.ToJSON(&pF)

		log.Printf("%+v", pF)

		pesan = strings.Replace(pesan, "{{" + url + "}}", pF.Message, -1)
		log.Println(pesan)
	}

	return pesan
}

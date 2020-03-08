package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

type SteamResponse struct {
	Steamid string
	Success int
}

type Steam struct {
	Response SteamResponse
}

func getData(url string) []byte {

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(body))
	return body
}

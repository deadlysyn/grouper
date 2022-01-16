package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/deadlysyn/retriever"
)

func slackNotify(msg string) error {
	creds, err := retriever.Fetch()
	if err != nil {
		return err
	}

	// json.Marshal feels right, but causes invalid payload from slack
	j := []byte(`{"text":"` + msg + `"}`)

	res, err := http.Post(creds["SLACK_WEBHOOK"], "application/json", bytes.NewBuffer(j))
	if res != nil {
		defer res.Body.Close()
		if res.StatusCode >= 400 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Printf("slack webhook http status: %d (%s)", res.StatusCode, body)
		}
	}
	if err != nil {
		return err
	}

	return nil
}

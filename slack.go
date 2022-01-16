package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func slackNotify(msg string) error {
	webhook := os.Getenv("SLACK_WEBHOOK")
	if len(webhook) == 0 {
		return errors.New("SLACK_WEBHOOK is not defined")
	}

	j := []byte(`{"text":"` + msg + `"}`)

	res, err := http.Post(webhook, "application/json", bytes.NewBuffer(j))
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

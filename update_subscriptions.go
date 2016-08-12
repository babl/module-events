package main

import (
	"io"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
)

const SubscriptionsApiUrl = "https://babl.sh/api/subscriptions"

func updateSubscriptions() {
	// TODO: timeout after 10 seconds
	resp, err := http.Get(SubscriptionsApiUrl)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Updating subscriptions failed")
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(SubscriptionsPath)
	check(err)
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	check(err)
	err = file.Sync()
	check(err)
}

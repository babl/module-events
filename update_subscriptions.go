package main

import (
	"io"
	"net/http"
	"os"
)

const SubscriptionsApiUrl = "https://babl.sh/api/subscriptions"

func updateSubscriptions() {
	resp, err := http.Get(SubscriptionsApiUrl)
	check(err)
	defer resp.Body.Close()

	file, err := os.Create(SubscriptionsPath)
	check(err)
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	check(err)
	err = file.Sync()
	check(err)
}

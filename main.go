//go:generate go-bindata -nocompress subscriptions.yml

package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type subscription struct {
	Exec string            `yaml:"exec"`
	Env  map[string]string `yaml:"env"`
}

type config map[string][]subscription

func main() {
	event := os.Getenv("EVENT")
	log.Printf("Event triggered: %s\n", event)

	contents, err := Asset("subscriptions.yml")
	check(err)

	var c config
	err = yaml.Unmarshal(contents, &c)
	check(err)

	for e, subs := range c {
		if e == event {
			for _, sub := range subs {
				log.Printf("EXEC babl --host queue.babl.sh --port 4445 %s %q", sub.Exec, sub.Env)
			}
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

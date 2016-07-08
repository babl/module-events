//go:generate go-bindata -nocompress subscriptions.yml

package main

import (
	"log"
	"os"
	"sync"

	babl "github.com/larskluge/babl/shared"
	"gopkg.in/yaml.v2"
)

type Subscription struct {
	Exec string   `yaml:"exec"`
	Env  babl.Env `yaml:"env"`
}

type config map[string][]Subscription

func main() {
	event := os.Getenv("EVENT")
	log.Printf("Event triggered: %s\n", event)

	contents, err := Asset("subscriptions.yml")
	check(err)
	var c config
	err = yaml.Unmarshal(contents, &c)
	check(err)

	stdin := babl.ReadStdin()

	var wg sync.WaitGroup
	for e, subs := range c {
		if e == event {
			for _, sub := range subs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					exec(sub.Exec, sub.Env, &stdin)
				}()
			}
		}
	}
	wg.Wait()
}

func exec(moduleName string, env babl.Env, stdin *[]byte) {
	log.Printf("EXEC babl --async %s %q", moduleName, env)
	module := babl.NewModule(moduleName)
	module.Address = "queue.babl.sh:4445"
	module.Env = env
	module.SetAsync(true)
	_, _, _, err := module.Call(*stdin)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

//go:generate go-bindata -nocompress subscriptions.yml

package main

import (
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/cenk/backoff"
	babl "github.com/larskluge/babl/shared"
	"gopkg.in/yaml.v2"
)

type Subscription struct {
	Exec string   `yaml:"exec"`
	Env  babl.Env `yaml:"env"`
}

type config map[string][]Subscription

func init() {
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	event := os.Getenv("EVENT")

	if event == "" {
		log.Warn("No EVENT given")
		os.Exit(0)
	}

	contents, err := Asset("subscriptions.yml")
	check(err)
	var c config
	err = yaml.Unmarshal(contents, &c)
	check(err)

	stdin := babl.ReadStdin()

	n := 0
	var wg sync.WaitGroup
	for e, subs := range c {
		if e == event {
			for _, sub := range subs {
				wg.Add(1)
				n += 1
				go func(sub Subscription) {
					defer wg.Done()
					fn := func() error {
						return exec(sub.Exec, sub.Env, &stdin)
					}
					backoff.Retry(fn, backoff.NewExponentialBackOff())
					if err != nil {
						log.WithError(err).Warn("Subscription could not be executed")
					}
				}(sub)
			}
		}
	}
	log.WithFields(log.Fields{"event": event, "subscriptions": n}).Info("Event Triggered")
	wg.Wait()
}

func exec(moduleName string, env babl.Env, stdin *[]byte) error {
	if env == nil {
		env = babl.Env{}
	}
	env = includeForwardedEnv(env)
	log.WithFields(log.Fields{"module": moduleName, "env": env}).Info("Executing Module")
	module := babl.NewModule(moduleName)
	module.Address = "queue.babl.sh:4445"
	module.Env = env
	module.SetAsync(true)
	_, _, _, err := module.Call(*stdin)
	return err
}

func includeForwardedEnv(env babl.Env) babl.Env {
	varList := os.Getenv("BABL_VARS")
	if varList != "" {
		vars := strings.Split(varList, ",")
		for _, k := range vars {
			if _, exists := env[k]; !exists { // do not overwrite subscription configuration values
				env[k] = os.Getenv(k)
			}
		}
	}
	return env
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

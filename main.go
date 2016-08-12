package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/cenk/backoff"
	"github.com/larskluge/babl/bablmodule"
	"github.com/larskluge/babl/bablutils"
)

const SubscriptionsPath = "subscriptions.json"

var updateSubscriptionsFlag = flag.Bool("update", false, "Update Subscriptions & exit")

type Subscription struct {
	Exec string         `json:"module"`
	Env  bablmodule.Env `json:"env"`
}

type config map[string][]Subscription

func init() {
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	flag.Parse()
	event := os.Getenv("EVENT")

	log.Info("0")

	if event == "babl:subscriptions:updated" || *updateSubscriptionsFlag {
		log.Info("Updating event subscriptions from babl.sh")
		updateSubscriptions()
		os.Exit(0)
	}
	log.Info("1")

	if event == "" {
		log.Warn("No EVENT given")
		os.Exit(0)
	}

	log.Info("before")
	contents, err := ioutil.ReadFile(SubscriptionsPath)
	log.Info("after")
	check(err)
	var c config
	log.Info("before1")
	err = json.Unmarshal(contents, &c)
	log.Info("after2")
	check(err)

	stdin := bablutils.ReadStdin()

	log.Info("4")

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

func exec(moduleName string, env bablmodule.Env, stdin *[]byte) error {
	if env == nil {
		env = bablmodule.Env{}
	}
	env = includeForwardedEnv(env)
	log.WithFields(log.Fields{"module": moduleName, "env": env}).Info("Executing Module")
	module := bablmodule.New(moduleName)
	module.Address = "queue.babl.sh:4445"
	// module.Address = "localhost:4445"
	module.Env = env
	module.SetAsync(true)
	module.SetDebug(true)
	_, _, stderr, err := module.Call(*stdin)
	log.Warn(stderr)
	return err
}

func includeForwardedEnv(env bablmodule.Env) bablmodule.Env {
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

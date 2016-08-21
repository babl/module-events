package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

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

	if event == "babl:subscriptions:updated" || *updateSubscriptionsFlag {
		log.Info("Updating event subscriptions from babl.sh")
		updateSubscriptions()
		os.Exit(0)
	}

	if event == "" {
		log.Warn("No EVENT given")
		os.Exit(0)
	}

	contents, err := ioutil.ReadFile(SubscriptionsPath)
	check(err)
	var c config
	err = json.Unmarshal(contents, &c)
	check(err)

	stdin := bablutils.ReadStdin()

	n := 0
	var wg sync.WaitGroup
	for e, subs := range c {
		if e == event {
			for _, sub := range subs {
				wg.Add(1)
				n += 1
				go func(sub Subscription) {
					defer wg.Done()
					operation := func() error {
						return exec(sub.Exec, sub.Env, &stdin)
					}
					notify := func(err error, duration time.Duration) {
						log.WithFields(log.Fields{"duration": duration}).WithError(err).Warn("babl/events: module exec failed, retrying..")
					}
					err := backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
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
	stdout, stderr, exitcode, err := module.Call(*stdin)
	if err != nil {
		log.WithFields(log.Fields{"stdout": string(stdout), "stderr": string(stderr), "exitcode": exitcode, "error": err, "module": moduleName, "env": env, "stdin": string(*stdin)}).Warn("babl/events: Module call failed")

		// unknown module requested to be triggered, ignoring
		if strings.Contains(err.Error(), "unknown service") {
			err = nil
		}
	}

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

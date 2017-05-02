package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aisola/log"
	"github.com/aisola/log/http"
	"gopkg.in/go-playground/webhooks.v3/github"
)

type Options struct {
	App *App
	Secret string
}

func NewHookHandler(options *Options) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		eventName := r.Header.Get("X-Github-Event")
		if eventName != "push" {
			log.Infof("Ignoring '%s' event", eventName)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			loghttp.Error(w, 500, "Internal Server Error", err)
			return
		}

		if options.Secret != "" {
			ok := false
			for _, sig := range strings.Fields(r.Header.Get("X-Hub-Signature")) {
				if !strings.HasPrefix(sig, "sha1=") {
					continue
				}
				sig = strings.TrimPrefix(sig, "sha1=")
				mac := hmac.New(sha1.New, []byte(options.Secret))
				mac.Write(body)
				if sig == hex.EncodeToString(mac.Sum(nil)) {
					ok = true
					break
				}
			}
			if !ok {
				log.Infof("Ignoring '%s' event with incorrect signature", eventName)
				return
			}
		}

		event := new(github.PushPayload)
		if err := json.Unmarshal(body, event); err != nil {
			log.Infof("Ignoring '%s' event with invalid payload", eventName)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		repoFullName := fmt.Sprintf("%s/%s", options.App.User, options.App.Name)
		if  event.Repository.FullName != repoFullName {
			log.Infof("Ignoring '%s' event with incorrect repository name", eventName)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		log.Infof("Handling '%s' event for %s", eventName, repoFullName)

		if err := options.App.Update(); err != nil {
			loghttp.Error(w, 500, "Internal Server Error", err)
			return
		}
	})
}

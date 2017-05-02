package main

import (
	"fmt"
	"net/http"

	"github.com/aisola/log"
	"github.com/pressly/chi"
	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("core.port", 2017)
	viper.SetDefault("core.repo", "$HOME/repository") // git repo
	viper.SetDefault("core.source", "$HOME/source")   // work tree / app dir
	viper.SetDefault("core.public", "$HOME/public")   // public_html root
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Info("Could not load in configuration...")
		return
	}
	log.DefaultLogger.SetDebug(true)

	r := chi.NewRouter()

	// Set up the application
	app := NewApp(viper.GetString("hugo.repo"))
	options := &Options{App: app, Secret: viper.GetString("core.secret")}
	r.Mount("/webhook", NewHookHandler(options))


	port := fmt.Sprintf(":%d", viper.GetInt("core.port"))
	log.Infof("Serving on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Info("Could not start server.")
		log.Debug(err.Error())
	}
}

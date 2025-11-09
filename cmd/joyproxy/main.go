// Package main represents entry point function of application.
package main

import (
	"net/http"
	"os"
	"time"

	"joyproxy/internal/joyproxy"
	"joyproxy/internal/log"
)

// main application entry point.
func main() {
	joyproxy.LoadConf()

	log.Init(joyproxy.Cfg.Loglevel, os.Stderr)

	http.HandleFunc("/joyproxy/", joyproxy.JoyproxyHandler)
	http.HandleFunc("/joyurl", joyproxy.JoyurlHandler)

	log.Infof("Starting application on port %v\n", joyproxy.Cfg.Port)

	// To prevent slow loris ddos, we have to define custom handler because builtin http.ListenAndServe() have no
	// timeout and sits here forever.
	server := &http.Server{
		Addr:              joyproxy.Cfg.Port,
		ReadHeaderTimeout: time.Duration(joyproxy.Cfg.Timeout) * time.Second,
		ReadTimeout:       time.Duration(joyproxy.Cfg.Timeout) * time.Second,
		WriteTimeout:      time.Duration(joyproxy.Cfg.Timeout) * time.Second,
		IdleTimeout:       time.Duration(joyproxy.Cfg.Timeout) * time.Second,
	}

	log.Warn(server.ListenAndServe().Error())
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

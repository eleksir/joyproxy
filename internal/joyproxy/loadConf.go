package joyproxy

import (
	"os"
	"path/filepath"

	"joyproxy/internal/log"

	"github.com/wlevene/ini"
)

// LoadConf reads config data from config file and parses it to Cfg variable.
func LoadConf() {
	executablePath, err := os.Executable()

	if err != nil {
		log.Fatalf("Unable to get current executable path: %s", err)
	}

	configINIPath := filepath.Dir(executablePath) + "/joyproxy.ini"
	configFH := ini.New().LoadFile(configINIPath)

	Cfg.Port = configFH.Section("settings").Get("port")

	if Cfg.Port == "" {
		log.Fatalf("unable to get port setting from settings section of %s", configINIPath)
	}

	Cfg.Port = ":" + Cfg.Port

	Cfg.Proto = configFH.Section("settings").Get("proto")

	if Cfg.Proto == "" {
		log.Fatalf("unable to get proto setting from settings section of %s", configINIPath)
	}

	if Cfg.Proto != "http" && Cfg.Proto != "https" {
		log.Fatalf("proto setting from settings section of %s must be either http or https, but we got %s",
			configINIPath,
			Cfg.Proto,
		)
	}

	Cfg.Timeout = configFH.Section("settings").GetInt("timeout")

	if Cfg.Timeout == 0 {
		log.Fatalf("unable to get timeout value from settings section of %s", configINIPath)
	}

	Cfg.Loglevel = configFH.Section("settings").Get("loglevel")

	if Cfg.Loglevel == "" {
		log.Fatalf("unable to get loglevel setting from settings section of %s", configINIPath)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

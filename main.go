package main

import (
	"fmt"
	"html"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type config struct {
	port     string
	proto    string
	timeout  int
	loglevel string
}

var cfg config

// Pretend, we're… isp nat or proxy6 so we have different user agent.
var userAgentString = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36 OPR/83.0.4254.27",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Vivaldi/3.5",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36 Edg/88.0.705.81",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36 Edg/95.0.1020.44",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36 Edg/96.0.1054.62",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36 OPR/82.0.4227.50",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36 Edg/97.0.1072.55",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36 Edg/97.0.1072.69",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36 Vivaldi/4.3",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 YaBrowser/22.1.0 Yowser/2.5 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:91.0) Gecko/20100101 Firefox/91.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:95.0) Gecko/20100101 Firefox/95.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:96.0) Gecko/20100101 Firefox/96.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:97.0) Gecko/20100101 Firefox/97.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Vivaldi/3.5",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36 OPR/83.0.4254.27",
	"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36 OPR/83.0.4254.27",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.81 Safari/537.36 Vivaldi/4.3",
	"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:91.0) Gecko/20100101 Firefox/91.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:97.0) Gecko/20100101 Firefox/97.0",
}

var postForm = "<html>\n<body>\n<form method='get' action='joyurl'>\n<input type='text' name='joyurl' size=100 autofocus><br>\n<input type='submit' value='Post it!' style='font-size:115%%;'>\n<br>%s<br>\n</body>\n</html>\n"

func loadConf() {
	executablePath, err := os.Executable()

	if err != nil {
		log.Errorf("Unable to get current executable path: %s", err)
		os.Exit(1)
	}

	configINIPath := fmt.Sprintf("%s/joyproxy.ini", filepath.Dir(executablePath))
	configFH, err := ini.Load(configINIPath)

	if err != nil {
		log.Errorf("Unable to read config file: %s", err)
		os.Exit(1)
	}

	cfg.port = configFH.Section("settings").Key("port").String()

	if cfg.port == "0" {
		log.Errorf("unable to get port setting from settings section of %s", configINIPath)
		os.Exit(1)
	}

	cfg.port = fmt.Sprintf(":%s", cfg.port)

	cfg.proto = configFH.Section("settings").Key("proto").String()

	if cfg.proto == "0" {
		log.Errorf("unable to get proto setting from settings section of %s", configINIPath)
		os.Exit(1)
	}

	if cfg.proto != "http" && cfg.proto != "https" {
		log.Errorf("proto setting from settings section of %s must be either http or https, but we got %s", configINIPath, cfg.proto)
		os.Exit(1)
	}

	cfg.timeout, err = configFH.Section("settings").Key("timeout").Int()

	if err != nil {
		log.Errorf("unable to get timeout value from settings section: %s", err)
		os.Exit(1)
	}

	cfg.loglevel = configFH.Section("settings").Key("loglevel").String()

	if cfg.loglevel == "0" {
		log.Errorf("unable to get loglevel setting from settings section of %s", configINIPath)
		os.Exit(1)
	}
}

func joyproxyHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.Method != http.MethodGet {
		log.Infof("request method is not GET, but %s", r.Method)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	dlReq := r
	joyPath := strings.Split(r.RequestURI, "/")
	// joyPath[0] is empty, joyPath[1] is joyproxy
	joyHost := joyPath[2]

	if !regexp.MustCompile("^img[0-9]+[.]reactor[.]cc$").MatchString(joyHost) {
		log.Infof("supplied proxy target host is not match with ^img[0-9]+[.]reactor[.]cc$ pattern: %s", joyHost)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	joyPath = joyPath[3:]
	dlReq.RequestURI = "/" + strings.Join(joyPath, "/")

	if !regexp.MustCompile("^/pics/post/mp4/.+[.]mp4$").MatchString(dlReq.RequestURI) {
		log.Infof("supplied uri is not match with /pics/post/mp4/.+[.]mp4$ pattern: %s", dlReq.RequestURI)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	dlReq.Host = joyHost
	dlReq.URL, err = url.Parse(dlReq.RequestURI)

	if err != nil {
		log.Infof("unable to parse %s via url.Parse(): %s", dlReq.RequestURI, err)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	// Forge request as it comes from browser… kinda
	if dlReq.Header.Get("Referer") != "" {
		dlReq.Header.Del("Referer")
	}

	dlReq.Header.Add("Referer", "https://old.reactor.cc/all")

	if dlReq.Header.Get("User-Agent") != "" {
		dlReq.Header.Del("User-Agent")
	}

	dlReq.Header.Add("User-Agent", userAgentString[rand.Intn(len(userAgentString)-1)])

	if dlReq.Header.Get("Accept") != "" {
		dlReq.Header.Del("Accept")
	}

	dlReq.Header.Add("Accept", "*/*")

	if dlReq.Header.Get("Accept-Encoding") != "" {
		dlReq.Header.Del("Accept-Encoding")
	}

	dlReq.Header.Add("Accept-Encoding", "identity;q=1, *;q=0")

	if dlReq.Header.Get("Accept-Language") != "" {
		dlReq.Header.Del("Accept-Language")
	}

	dlReq.Header.Add("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	if dlReq.Header.Get("Range") != "" {
		dlReq.Header.Del("Range")
	}

	dlReq.Header.Add("Range", "bytes=0-")
	log.Debugf(spew.Sdump(dlReq))

	u := fmt.Sprintf("https://%s", joyHost)
	target, err := url.Parse(u)

	if err != nil {
		log.Infof("unable to parse %s via url.Parse(): %s", u, err)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	joyproxy := httputil.NewSingleHostReverseProxy(target)
	joyproxy.ServeHTTP(w, dlReq)
}

func joyurlHandler(w http.ResponseWriter, r *http.Request) {
	var (
		param          = r.URL.Query()
		proxyURLString = " "
	)

	log.Debugln(spew.Sdump(param))

	re := regexp.MustCompile("^https?://img[0-9]+[.]reactor[.]cc/.+[.][Ww][Ee][Bb][Mm]$")

	if len(param) > 0 && re.MatchString(param["joyurl"][0]) {
		log.Debug("match with ^https?://img[0-9]+[.]reactor[.]cc/.+[.][Ww][Ee][Bb][Mm]$ pattern")
		// https://img1.reactor.cc/pics/post/webm/видосик.webm

		p := regexp.MustCompile("/").Split(param["joyurl"][0], -1)
		log.Debug(spew.Sdump(p))

		if len(p) > 6 && p[5] == "webm" {
			log.Debug("webm url part detected, let's make joyproxy url")

			file := p[6][:len(p[6])-5]

			proxyURLString = fmt.Sprintf(
				"%s://%s/joyproxy/%s/%s/%s/mp4/%s.mp4",
				cfg.proto,
				r.Host,
				p[2],
				p[3],
				p[4],
				file,
			)

			proxyURLString = html.EscapeString(proxyURLString)
		} else {
			log.Debug("webm url part not detected, skip making joyproxy url")
		}
	} else {
		log.Info("not match with ^https?://img[0-9]+[.]reactor[.]cc/.+[.][Ww][Ee][Bb][Mm]$ regex")
	}

	htmlText := fmt.Sprintf(postForm, proxyURLString)

	if _, err := fmt.Fprintln(w, htmlText); err != nil {
		log.Infof("Unable to write response to client: %s", err)
	}
}

func main() {
	// setup logging
	log.SetFormatter(&log.TextFormatter{
		DisableQuote:           true,
		DisableLevelTruncation: false,
		DisableColors:          true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
	})

	loadConf()
	// info or debug, default info
	switch cfg.loglevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	http.HandleFunc("/joyproxy/", joyproxyHandler)
	http.HandleFunc("/joyurl", joyurlHandler)

	log.Infof("Starting application on port %v\n", cfg.port)

	// To prevent slow loris ddos, we have to define custom handler because builtin http.ListenAndServe() have no
	// timeout and sits here forever.
	server := &http.Server{
		Addr:              cfg.port,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Warn(server.ListenAndServe())
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

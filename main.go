package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"net/http"

	"github.com/goji/httpauth"
	stratum_ping "github.com/xunzhou/stratum-ping"
	"gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
	"gopkg.in/yaml.v2"
)

type Server struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol,omitempty"`
	TLS      bool   `yaml:"tls,omitempty"`
}

type Config struct {
	Servers []Server `yaml:"servers"`
	Cred    struct {
		User   string `yaml:"user"`
		Passwd string `yaml:"passwd"`
	} `yaml:"cred"`
	Tls struct {
		Cert string `yaml:"cert"`
		Priv string `yaml:"key"`
	} `yaml:"tls",omitempty`
	Apiproxy struct {
		Enable bool   `yaml:"enable"`
		Miner  string `yaml:"miner"`
	} `yaml:"ethermine-api-proxy",omitempty`
}

var config Config
var PORT = ":3001"
var TLSPORT = ":8443"

func cli_ping() {
	argLogin := flag.String("u", "0x0000000000000000000000000000000000000000", "login")
	argPass := flag.String("p", "x", "pass")
	argCount := flag.Int("c", 4, "stop after <count> replies")
	argV6 := flag.Bool("6", false, "use ipv6")
	argProto := flag.String("t", "stratum2", "stratum type: stratum1, stratum2")
	argTLS := flag.Bool("tls", false, "use TLS")

	flag.Parse()

	argServer := flag.Arg(0)

	if len(argServer) == 0 {
		fmt.Printf("Stratum server cannot be empty\n\n")
		return
	}

	split := strings.Split(argServer, ":")
	if len(split) != 2 {
		fmt.Printf("Invalid host/port specified\n\n")
		return
	}

	if *argCount <= 0 || *argCount > 20000 {
		fmt.Printf("Invalid count specified\n\n")
		return
	}

	portNum, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil || portNum <= 0 || portNum >= 65536 {
		fmt.Printf("Invalid port specified\n\n")
		return
	}

	switch *argProto {
	case "stratum1":
		fallthrough
	case "stratum2":
		break
	default:
		fmt.Printf("Invalid stratum type specified\n\n")
		return
	}

	pinger := stratum_ping.StratumPinger{
		Login: *argLogin,
		Pass:  *argPass,
		Count: *argCount,
		Host:  split[0],
		Port:  split[1],
		Ipv6:  *argV6,
		Proto: *argProto,
		Tls:   *argTLS,
	}

	fmt.Println(pinger.Do())
}

func ping(proto, host string, port int, tls bool, ch chan stratum_ping.Result) stratum_ping.Result {
	pinger := stratum_ping.StratumPinger{
		Login: "0x0000",
		Pass:  "",
		Count: 5,
		Host:  host,
		Port:  strconv.Itoa(port),
		Ipv6:  false,
		Proto: proto,
		Tls:   tls,
	}
	res := pinger.Do()
	ch <- res
	return res
}

func logging(r *http.Request) {
	log.Printf("%s %s %s %s\n", r.RemoteAddr, r.Method, r.URL, r.UserAgent())
}

func status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
	logging(r)
}

func api(w http.ResponseWriter, r *http.Request) {
	uri := strings.Replace(r.RequestURI, "/api", "", 1)
	fmt.Fprint(w, apireq(uri))
	logging(r)
}

func all(w http.ResponseWriter, r *http.Request) {
	res := pingAll()
	fmt.Fprintf(w, "%s", res)
	logging(r)
}

func apireq(uri string) string {
	resp, err := http.Get("https://api.ethermine.org" + uri)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	log.Printf("Remote API Server >> GET https://api.ethermine.org" + uri)

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}

var router = mux.NewRouter()

func HandlerFunc(p string, f func(http.ResponseWriter, *http.Request)) {
	router.HandleFunc(p, f).Methods("GET")
	http.Handle(p, httpauth.SimpleBasicAuth(config.Cred.User, config.Cred.Passwd)(router))
}

func handleRequests() {
	HandlerFunc("/", status)
	HandlerFunc("/all", all)
	if config.Apiproxy.Enable {
		miner := config.Apiproxy.Miner
		HandlerFunc("/api", api)
		HandlerFunc("/api/{.*}stats", api)
		HandlerFunc("/api/miner/"+miner+"/dashboard", api)
		HandlerFunc("/api/miner/"+miner+"/dashboard/payouts", api)
		HandlerFunc("/api/miner/"+miner+"/worker/{worker}/history", api)
	}

	if len(config.Tls.Cert) == 0 || len(config.Tls.Priv) == 0 {
		log.Println("Listening on " + PORT)
		log.Fatal(http.ListenAndServe(PORT, nil))
	} else {
		log.Println("Listening on TLS " + TLSPORT)
		log.Fatal(http.ListenAndServeTLS(TLSPORT, config.Tls.Cert, config.Tls.Priv, nil))
	}

}

func loadConfig() {
	data, err := ioutil.ReadFile("stratum-health.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

func pingAll() string {
	// res := ""
	res := []stratum_ping.Result{}
	ch := make(chan stratum_ping.Result)

	for _, s := range config.Servers {
		if s.Protocol == "" {
			s.Protocol = "stratum2"
		}
		go ping(s.Protocol, s.Host, s.Port, s.TLS, ch)
	}

	for i := 0; i < len(config.Servers); i++ {
		pingRes := <-ch
		res = append(res, pingRes)
	}

	j, _ := json.Marshal(res)
	return string(j)
}

func main() {
	if len(os.Args) > 1 {
		cli_ping()
	} else {
		loadConfig()
		// fmt.Println(pingAll())
	}

	handleRequests()
}

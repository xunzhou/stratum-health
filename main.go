package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"net/http"

	stratum_ping "github.com/xunzhou/stratum-ping"
	"gopkg.in/yaml.v2"
)

type Server struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `default:"stratum2" yaml:"protocol",omitempty`
	TLS      bool   `default:false yaml:"tls",omitempty`
}

type Servers struct {
	List []Server `yaml:"servers"`
}

var servers Servers
var PORT = ":3001"

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

	// if err := pinger.Do(); err != nil {
	// 	fmt.Printf("%s\n\n", err)
	// }
	fmt.Println(pinger.Do())
}

func ping(proto, host string, port int, tls bool) string {
	pinger := stratum_ping.StratumPinger{
		Login: "0x0000",
		Pass:  "",
		Count: 5,
		Host:  host,
		Port:  strconv.Itoa(port),
		Ipv6:  false,
		Proto: "stratum2",
		Tls:   tls,
	}
	res := pinger.Do()
	return res
}

func logging(r *http.Request) {
	log.Printf("%s %s %s %s\n", r.RemoteAddr, r.Method, r.URL, r.UserAgent())
}

func status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
	logging(r)
}

func all(w http.ResponseWriter, r *http.Request) {
	res := pingAll()
	fmt.Fprintf(w, "%s", res)
	logging(r)
}

func handleRequests() {
	http.HandleFunc("/", status)
	http.HandleFunc("/all", all)
	log.Println("Listening on" + PORT)
	log.Fatal(http.ListenAndServe(PORT, nil))
}

func loadConfig() {
	data, err := ioutil.ReadFile("stratum-health.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &servers)
	if err != nil {
		panic(err)
	}
}

func pingAll() string {
	res := ""
	for _, s := range servers.List {
		if s.Protocol == "" {
			s.Protocol = "stratum2"
		}
		// fmt.Printf("[TLS:%t] %s://%s:%d\n", s.TLS, s.Protocol, s.Host, s.Port)
		res += fmt.Sprintf("%s:%d\n%s\n", s.Host, s.Port, ping(s.Protocol, s.Host, s.Port, s.TLS))
	}
	return res
}

func main() {
	if len(os.Args) > 1 {
		cli_ping()
	} else {
		loadConfig()
		// pingAll()
	}

	handleRequests()
}

package main

import (
	"flag"
	"fmt"
	"simpleauth"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"oauth"
	"runtime"
	"scenario"
	"strings"
	"sync/atomic"
	"time"
)

// to reduce size of thread, speed up
const SizePerThread = 10000000

// for proxy, debug use only
//var DefaultTransport RoundTripper = &Transport{Proxy: ProxyFromEnvironment}

// Counter will be an atomic, to count the number of request handled
// which will be used to print PPS, etc.
type Counter struct {
	totalReq     int64 // total # of request
	totalResTime int64 // total response time
	totalErr     int64 // how many error
	totalResSlow int64 // how many slow response
	totalSend    int64

	lastSend int64
	lastReq  int64

	client  *http.Client
	monitor *time.Ticker
	// ideally error should be organized by type TODO
	throttle <-chan time.Time
}

// init
func (c *Counter) Init() {
	// set up HTTP proxy
	if proxy != "none" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		c.client = &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 200000,
				Proxy:               http.ProxyURL(proxyUrl),
			},
		}
	} else {
		c.client = &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 200000,
			},
		}
	}
}

// increase the count and record response time.
func (c *Counter) recordRes(_time int64, method string) {
	atomic.AddInt64(&c.totalReq, 1)
	atomic.AddInt64(&c.totalResTime, _time)

	// if longer that 200ms, it is a slow response
	if _time > slowThreshold*1000000 {
		atomic.AddInt64(&c.totalResSlow, 1)
		log.Println("slow response -> ", float64(_time)/1.0e9, method)
	}
}

func (c *Counter) recordError() {
	atomic.AddInt64(&c.totalErr, 1)
}

func (c *Counter) recordSend() {
	atomic.AddInt64(&c.totalSend, 1)
}

// main goroutine to drive traffic
func (c *Counter) hammer(rg *rand.Rand) {
	// before send out, update send count
	c.recordSend()
	call, err := profile.NextCall(rg)

	if err != nil {
		log.Println("next call error: ", err)
		return
	}

	req, err := http.NewRequest(call.Method, call.URL, strings.NewReader(call.Body))
	// log.Println(call, req, err)
	switch auth_method {
	case "oauth":
		_signature := oauth_client.AuthorizationHeaderWithBodyHash(nil, call.Method, call.URL, url.Values{}, call.Body)
		req.Header.Add("Authorization", _signature)
	case "simples2s":
		// simple authen here
		// SignS2SRequest(method string, url string, body string) (string, string, error)
		_signature, _timestamp, _ := simple_client.SignS2SRequest(call.Method, call.URL, call.Body)

		req.Header.Add("Authorization", "S2S"+" realm=\"modern-war\""+
			", signature=\""+_signature+"\", timestamp=\""+_timestamp+"\"")
	case "intrenalc2s":
		// simple authen here
		// SignS2SRequest(method string, url string, body string) (string, string, error)
		_signature, _timestamp, _ := simple_client.SignC2SRequest(call.Method, call.URL, call.Body)

		req.Header.Add("Authorization", "C2S"+" realm=\"jackpot-slots\""+
			", signature=\""+_signature+"\", timestamp=\""+_timestamp+"\"")
	}

	// Add special haeader for PATCH, PUT and POST
	switch call.Method {
	case "PATCH", "PUT", "POST":
		switch call.Type {
		case "REST":
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			break
		case "WWW":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			break
		}
	}

	t1 := time.Now().UnixNano()
	res, err := c.client.Do(req)

	response_time := time.Now().UnixNano() - t1

	/*
		    ###
			disable reading res.body, no need for our purpose for now,
		    by doing this, hope we can save more file descriptor.
			##
	*/
	defer req.Body.Close()

	switch {
	case err != nil:
		log.Println("Response Time: ", float64(response_time)/1.0e9, " Erorr: when", call.Method, call.URL, "with error ", err)
		c.recordError()
	case res.StatusCode >= 400 && res.StatusCode != 409:
		log.Println("Got error code --> ", res.Status, "for call ", call.Method, " ", call.URL)
		c.recordError()
	default:
		// only do successful response here
		defer res.Body.Close()
		c.recordRes(response_time, call.URL)
		data, _ := ioutil.ReadAll(res.Body)
		if call.CallBack == nil && !debug {
		} else {
			if res.StatusCode == 409 {
				log.Println("Http 409 Res Body : ", string(data))
			}
			if debug {
				log.Println("Req : ", call.Method, call.URL)
				if auth_method != "none" {
					log.Println("Authorization: ", string(req.Header.Get("Authorization")))
				}
				log.Println("Req Body : ", call.Body)
				log.Println("Response: ", res.Status)
				log.Println("Res Body : ", string(data))
			}
			if call.CallBack != nil {
				call.CallBack(call.SePoint, scenario.NEXT, data)
			}
		}
	}

}

func (c *Counter) monitorHammer() {
	sps := c.totalSend - c.lastSend
	pps := c.totalReq - c.lastReq
	backlog := c.totalSend - c.totalReq - c.totalErr

	atomic.StoreInt64(&c.lastReq, c.totalReq)
	atomic.StoreInt64(&c.lastSend, c.totalSend)

	avgT := float64(c.totalResTime) / (float64(c.totalReq) * 1.0e9)

	log.Println(
		" total: ", fmt.Sprintf("%4d", c.totalSend),
		" req/s: ", fmt.Sprintf("%4d", sps),
		" res/s: ", fmt.Sprintf("%4d", pps),
		" avg: ", fmt.Sprintf("%2.4f", avgT),
		" pending: ", backlog,
		" err:", c.totalErr,
		"|", fmt.Sprintf("%2.2f%s", (float64(c.totalErr)*100.0/float64(c.totalErr+c.totalReq)), "%"),
		" slow: ", fmt.Sprintf("%2.2f%s", (float64(c.totalResSlow)*100.0/float64(c.totalReq)), "%"),
		profile.CustomizedReport())
}

func (c *Counter) launch(rps int64) {
	_p := time.Duration(rps)
	_interval := 1000000000.0 / _p
	c.throttle = time.Tick(_interval * time.Nanosecond)

	go func() {
		i := 0
		for {
			if i == len(rands) {
				i = 0
			}
			<-c.throttle

			go c.hammer(rands[i])
			i++
		}
	}()

	c.monitor = time.NewTicker(time.Second)
	go func() {
		for {
			<-c.monitor.C // rate limit for monitor routine
			go c.monitorHammer()
		}
	}()
}

// init the program from command line
var (
	rps           int64
	profileFile   string
	profileType   string
	slowThreshold int64
	debug         bool
	auth_method   string
	sessionAmount int
	proxy         string

	// profile
	profile scenario.Profile

	// rands
	rands []*rand.Rand

	simple_client  = new(simpleauth.Client)
	oauth_client = new(oauth.Client)
)

func init() {
	flag.Int64Var(&rps, "rps", 500, "Set Request Per Second")
	flag.StringVar(&profileFile, "profile", "", "The path to the traffic profile")
	flag.Int64Var(&slowThreshold, "threshold", 200, "Set slowness standard (in millisecond)")
	flag.StringVar(&profileType, "type", "default", "Profile type (default|session|your session type)")
	flag.BoolVar(&debug, "debug", false, "debug flag (true|false)")
	flag.StringVar(&auth_method, "auth", "none", "Set authorization flag (oauth|simple(c|s)2s|none)")
	flag.IntVar(&sessionAmount, "size", 100, "session amount")
	flag.StringVar(&proxy, "proxy", "none", "Set HTTP proxy (need to specify scheme. e.g. http://127.0.0.1:8888)")

	simple_client.C2S_Secret = "----"
	simple_client.S2S_Secret = "----"
}

// main func
func main() {

	flag.Parse()
	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(1)

	// to speed up
	rands = make([]*rand.Rand, NCPU)
	for i, _ := range rands {
		s := rand.NewSource(time.Now().UnixNano())
		rands[i] = rand.New(s)
	}

	log.Println("cpu number -> ", NCPU)
	log.Println("rps -> ", rps)
	log.Println("slow threshold -> ", slowThreshold, "ms")
	log.Println("profile type -> ", profileType)
	log.Println("Proxy -> ", proxy)

	profile, _ = scenario.New(profileType, sessionAmount)
	if profileFile != "" {
		profile.InitFromFile(profileFile)
	} else {
		profile.InitFromCode()
	}

	rand.Seed(time.Now().UnixNano())

	counter := new(Counter)
	counter.Init()

	go counter.launch(rps)

	var input string
	fmt.Scanln(&input)
}

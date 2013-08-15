package main

// a simple mockup server to server WWW & RESTful API, used for quick test of hammer client

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)


func printHttpRequest(req *http.Request) {

	fmt.Println("\n")
	log.Println(req.Method, "  ", req.Host, req.URL)
	log.Println("Cookie: ", req.Cookies())
	log.Println("Request Headers : ")
	for key, value := range req.Header {
		log.Println("  ", key, " : ", value)
	}
	log.Println("Request Body : ")
	data, _ := ioutil.ReadAll(req.Body)

	log.Println(string(data), "\n")

}

// simple HTML Route
func hello(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(
		"Content-Type",
		"text/html",
	)
	io.WriteString(
		res,
		`<doctype html>
<html>
     <head>
           <title>Hello World</title>
     </head>
     <body>
           Hello World!
     </body>
</html>`,
	)
}

// a simple REST API
func hello_in_json(res http.ResponseWriter, req *http.Request) {

	printHttpRequest(req)

	res.Header().Set(
		"Content-Type",
		"text/json",
	)
	res.Header().Set(
		"My Header",
		"lol",
	)

	io.WriteString(
		res,
		`{"msg": "hello world",
    , timestamp: "nothing here"}`,
	)
}

func main() {
	helpString :=
		`	HTTP Port : 9000
	 /html : simple html response
	 /json : simple json response`


	fmt.Println(helpString)
	runtime.GOMAXPROCS(runtime.NumCPU())

	http.HandleFunc("/html", hello)
	http.HandleFunc("/json", hello_in_json)
	http.ListenAndServe(":9000", nil)
}

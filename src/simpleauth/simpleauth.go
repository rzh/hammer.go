package simpleauth

// used for lighter authentication of request

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client represents an OAuth client.
type Client struct {
	C2S_Secret string
	S2S_Secret string
}

var (
	S2S_Missing_Secret = errors.New("Server to Server secret is empty")
	C2S_Missing_Secret = errors.New("Client to Server secret is empty")
)

func (c *Client) signature(key string, method string, urlStr string, body string, timestamp string) string {
	w := hmac.New(sha1.New, []byte(key))
	// Method
	w.Write([]byte(strings.ToUpper(method)))

	// URL
	u, _ := url.Parse(urlStr)
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)
	path := u.RequestURI()
	switch {
	case scheme == "http" && strings.HasSuffix(host, ":80"):
		host = host[:len(host)-len(":80")]
	case scheme == "https" && strings.HasSuffix(host, ":443"):
		host = host[:len(host)-len(":443")]
	}

	w.Write([]byte(scheme + "://" + host + path))

	// write body
	w.Write([]byte(body))

	// write timestamp
	w.Write([]byte(timestamp))

	sum := w.Sum(nil)

	return hex.EncodeToString(sum)
}

/*
Sign the request for server 2 server call
 return value:
 	(signature, timestamp, error)
*/
func (c *Client) SignS2SRequest(method string, url string, body string) (string, string, error) {
	if c.S2S_Secret == "" {
		// no S2S sercret, return error
		return "", "", S2S_Missing_Secret
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	return c.signature(c.S2S_Secret, method, url, body, timestamp), timestamp, nil
}

func (c *Client) SignC2SRequest(method string, url string, body string) (string, string, error) {
	if c.S2S_Secret == "" {
		// no S2S sercret, return error
		return "", "", C2S_Missing_Secret
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	return c.signature(c.C2S_Secret, method, url, body, timestamp), timestamp, nil
}

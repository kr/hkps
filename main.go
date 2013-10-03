// Command ps is an hk plugin to list processes.
package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/kr/heroku-go"
)

// actual address to dial for network request
var dialAddr string

var tlsConfig = &tls.Config{
	InsecureSkipVerify: true,
}

func main() {
	log.SetFlags(0)
	maybePrintInfo()

	u := strings.TrimRight(getenv("WEBX_URL"), "/")
	if u == "" {
		log.Print("Please install webx.")
		log.Fatal("See https://devcenter.heroku.com/articles/webx")
	}
	req, err := newRequest("GET", u+"/mon/ps", nil)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{Transport: &http.Transport{Dial: dial,
		TLSClientConfig: tlsConfig}}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Println("status", resp.Status)
		io.Copy(os.Stderr, resp.Body)
		os.Exit(1)
	}
	var x []struct {
		Code   int
		Body   []byte
		Header http.Header
	}
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		log.Fatal(err)
	}

	printHeader := true
	for _, v := range x {
		dyno := v.Header.Get("Dyno")
		if v.Code != 200 {
			log.Println(dyno, strings.TrimSpace(string(v.Body)))
			continue
		}
		sc := bufio.NewScanner(bytes.NewReader(v.Body))
		if sc.Scan() && printHeader {
			fmt.Printf("%-8s %s\n", "DYNO", sc.Text())
			printHeader = false
		}
		for sc.Scan() {
			fmt.Printf("%-8s %s\n", dyno, sc.Text())
		}
	}
}

func getenv(name string) string {
	env := make(map[string]string)
	heroku.Get(&env, "/apps/"+os.Getenv("HKAPP")+"/config-vars")
	return env[name]
}

func dial(network, _ string) (net.Conn, error) {
	return net.Dial(network, dialAddr)
}

func newRequest(method, rawurl string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, rawurl, body)
	if err != nil {
		return nil, err
	}
	dialAddr = req.URL.Host
	if !strings.Contains(dialAddr, ":") {
		dialAddr += ":https"
	}
	req.Host = "api.webx.io"
	req.URL.Host = req.Host
	if req.URL.User == nil {
		return nil, errors.New("missing basic auth")
	}
	pw, _ := req.URL.User.Password()
	req.SetBasicAuth(req.URL.User.Username(), pw)
	req.URL.User = nil
	return req, nil
}

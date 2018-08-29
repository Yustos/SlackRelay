package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var config struct {
	BaseUpstreamURL *url.URL
	ListenAddress   string
}

func init() {
	// Init upstream URL
	envBaseURL := os.Getenv("SR_UPSTREAM_URL")
	if envBaseURL == "" {
		log.Fatal("Upstream URL not set.")
	}

	baseURL, err := url.Parse(envBaseURL)
	if err != nil {
		log.Fatalf("Incorrect upstream url: %v.", err)
	}
	config.BaseUpstreamURL = baseURL

	// Init listen address
	config.ListenAddress = os.Getenv("SR_LISTEN_ADDRESS")
	if config.ListenAddress == "" {
		log.Fatal("Listen address not set.")
	}

	log.Printf("Upstream URL: %v", config.BaseUpstreamURL)
	log.Printf("Listening on: %v", config.ListenAddress)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(config.ListenAddress, nil)
}

func handler(response http.ResponseWriter, request *http.Request) {
	log.Printf("Request: %s", request.RequestURI)

	upstreamResponse, err := sendMessage(request.URL, request.Body)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	log.Printf("Response: %s", upstreamResponse.Status)

	response.WriteHeader(upstreamResponse.StatusCode)

	// I don't like to read body
	defer upstreamResponse.Body.Close()
	upstreamBody, err := ioutil.ReadAll(upstreamResponse.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
	} else {
		response.Write(upstreamBody)
	}
}

func sendMessage(requestURL *url.URL, body io.ReadCloser) (*http.Response, error) {
	relURL, err := url.Parse(requestURL.RequestURI())
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	fullURL := config.BaseUpstreamURL.ResolveReference(relURL)

	return client.Post(fullURL.String(), "application/json", body)
}

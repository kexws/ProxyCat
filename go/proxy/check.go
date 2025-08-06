package proxy

import (
	"net/http"
	"net/url"
	"time"
)

// HTTPCheck verifies a proxy by requesting example.com through it.
func HTTPCheck(address string) bool {
	proxyURL, err := url.Parse(address)
	if err != nil {
		return false
	}
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	resp, err := client.Get("http://example.com")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

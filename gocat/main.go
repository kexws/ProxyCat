package main

import (
	"bufio"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Proxy struct {
	Address string `json:"address"`
}

var (
	validProxies  []Proxy
	failedProxies []Proxy
	mu            sync.Mutex
)

func loadProxies(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		addr := strings.TrimSpace(scanner.Text())
		if addr == "" {
			continue
		}
		p := Proxy{Address: addr}
		if checkProxy(addr) {
			validProxies = append(validProxies, p)
		} else {
			failedProxies = append(failedProxies, p)
		}
	}
}

func checkProxy(address string) bool {
	proxyURL, err := url.Parse(address)
	if err != nil {
		return false
	}
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		Timeout:   5 * time.Second,
	}
	resp, err := client.Get("https://httpbin.org/ip")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getValid(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	c.JSON(http.StatusOK, validProxies)
}

func getFailed(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	c.JSON(http.StatusOK, failedProxies)
}

type proxyRequest struct {
	Address string `json:"address"`
}

func recheckProxy(c *gin.Context) {
	var req proxyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if checkProxy(req.Address) {
		// remove from failed if exists
		for i, p := range failedProxies {
			if p.Address == req.Address {
				failedProxies = append(failedProxies[:i], failedProxies[i+1:]...)
				break
			}
		}
		// ensure not duplicated
		for _, p := range validProxies {
			if p.Address == req.Address {
				c.JSON(http.StatusOK, gin.H{"status": "already valid"})
				return
			}
		}
		validProxies = append(validProxies, Proxy{Address: req.Address})
		c.JSON(http.StatusOK, gin.H{"status": "restored"})
	} else {
		// ensure present in failed
		present := false
		for _, p := range failedProxies {
			if p.Address == req.Address {
				present = true
				break
			}
		}
		if !present {
			failedProxies = append(failedProxies, Proxy{Address: req.Address})
		}
		c.JSON(http.StatusOK, gin.H{"status": "still failed"})
	}
}

func deleteFailed(c *gin.Context) {
	var req proxyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	for i, p := range failedProxies {
		if p.Address == req.Address {
			failedProxies = append(failedProxies[:i], failedProxies[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"status": "deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func clearFailed(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	failedProxies = nil
	c.JSON(http.StatusOK, gin.H{"status": "cleared"})
}

func main() {
	loadProxies("../config/ip.txt")
	r := gin.Default()
	r.GET("/proxies/valid", getValid)
	r.GET("/proxies/failed", getFailed)
	r.POST("/proxies/recheck", recheckProxy)
	r.POST("/proxies/failed/delete", deleteFailed)
	r.DELETE("/proxies/failed", clearFailed)
	r.Run(":8080")
}

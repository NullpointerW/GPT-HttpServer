package gptcli

import (
	"gpt3.5/cfg"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NullpointerW/go-openai"
)

type Token struct {
	Context  []openai.ChatCompletionMessage
	LastTime time.Time
}

var (
	client       = newCli()
	TokenManager = sync.Map{}
)

func Cli() *openai.Client {
	return client.Load()
}

func tokensCleaner(d time.Duration) {
sleep:
	for {
		time.Sleep(d)
		goto clean
	}
clean:
	log.Println("token cleaner start")
	its := func(k, v interface{}) bool {
		tk := v.(*Token)
		if time.Since(tk.LastTime) >= d {
			TokenManager.Delete(k)
			log.Printf("clean token %s", k.(string))
		}
		return true
	}
	TokenManager.Range(its)
	goto sleep
}
func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	go tokensCleaner(time.Second * time.Duration(cfg.Cfg.TokenTTL))
}

func newCli() *atomic.Pointer[openai.Client] {
	cli := &atomic.Pointer[openai.Client]{}
	config := openai.DefaultConfig(cfg.Cfg.Apikey)
	transport := &http.Transport{}
	if cfg.Cfg.Proxy != "" {
		proxyUrl, err := url.Parse("http://" + cfg.Cfg.Proxy)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Cfg.Timeout) * time.Second,
	}
	cli.Store(openai.NewClientWithConfig(config))
	return cli
}

func SwitchCliWithApiKey(k string) {
	config := openai.DefaultConfig(k)
	transport := &http.Transport{}
	if cfg.Cfg.Proxy != "" {
		proxyUrl, err := url.Parse("http://" + cfg.Cfg.Proxy)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Cfg.Timeout) * time.Second,
	}
	client.Store(openai.NewClientWithConfig(config))
}

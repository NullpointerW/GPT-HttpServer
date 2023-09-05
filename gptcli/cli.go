package gptcli

import (
	"gpt-http/cfg"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NullpointerW/go-openai"
)

type PromptContext struct {
	Context  []openai.ChatCompletionMessage
	LastTime time.Time
}

var (
	client           = newCli()
	TokenManager     = sync.Map{}
	FineTunesManager = sync.Map{}
)

func Cli() *openai.Client {
	return client.Load()
}

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	go tokensCleaner(time.Second * time.Duration(cfg.Cfg.TokenTTL))
	go fineTunesCleaner(time.Hour * 24 * 2) // 2 days
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

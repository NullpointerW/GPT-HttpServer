package gptcli

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Token struct {
	Context  []openai.ChatCompletionMessage
	LastTime time.Time
}

var (
	apiKey       = "???"
	Cli          = newCli()
	TokenManager = sync.Map{}
)

func tokensCleaner(d time.Duration) {
	log.Println("token cleaner start")
clean:
	its := func(k, v interface{}) bool {
		tk := v.(*Token)
		if time.Since(tk.LastTime)>= d {
			TokenManager.Delete(k)
			log.Printf("clean token %s", k.(string))
		}
		return true
	}
	TokenManager.Range(its)

	for {
		time.Sleep(d)
		goto clean
	}
}
func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	go tokensCleaner(time.Minute * 30)
}

func newCli() *openai.Client {

	config := openai.DefaultConfig(apiKey)
	proxyUrl, err := url.Parse("http://localhost:7890")
	if err != nil {
		panic(err)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
		Timeout:   1 * time.Minute,
	}
	return openai.NewClientWithConfig(config)
}

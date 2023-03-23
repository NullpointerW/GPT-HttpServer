package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	// "github.com/go-playground/locales/lg"
	"github.com/sashabaranov/go-openai"
	"gpt3.5/cfg"
	"gpt3.5/gptcli"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Test website",
		})
	})

	r.GET("/stream", func(c *gin.Context) {

		var reqParam request
		err := c.ShouldBindQuery(&reqParam)
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Printf("reqpram is%#v \n", reqParam)

		if reqParam.Token == "" {
			//为了逐字显示
			msg := []rune("未获取到token,请检查token设置!")
			idx := 0
			len := len(msg)
			c.Stream(func(w io.Writer) bool {
				time.Sleep(100*time.Nanosecond)
				c.SSEvent("message", string(msg[idx]))
				if idx >= len-1 {
					return false
				}
				idx++
				return true
			})
			return
		}

		var req openai.ChatCompletionRequest

		if v, exist := gptcli.TokenManager.Load(reqParam.Token); exist {
			t := v.(*gptcli.Token)
			// fmt.Printf("token is %+v", t)
			tCtx := append(t.Context, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: reqParam.Message,
			})
			req = openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: tCtx,
			}
			ok, err, asw := SSEventProcess(c, req)
			if err != nil {
				log.Println(err)
			} else if ok {
				tCtx = append(tCtx, asw)
				t.Context = tCtx
				t.LastTime = time.Now()
			} else {
				log.Printf("send err :%#+v", reqParam)
			}
		} else { //not exist

			newCtx := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: cfg.Cfg.CharacterSetting,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: reqParam.Message,
				},
			}
			req = openai.ChatCompletionRequest{
				Model: openai.GPT3Dot5Turbo,
				//token
				Messages: newCtx,
			}
			ok, err, asw := SSEventProcess(c, req)
			if err != nil {
				log.Println(err)
			} else if ok {
				newCtx = append(newCtx, asw)
				gptcli.TokenManager.Store(reqParam.Token, &gptcli.Token{
					Context:  newCtx,
					LastTime: time.Now(),
				})
			} else {
				log.Printf("send err :%#+v", reqParam)
			}
		}

	})
	return r
}

func SSEventProcess(c *gin.Context, req openai.ChatCompletionRequest) (bool, error, openai.ChatCompletionMessage) {
	stream, err := gptcli.Cli().CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		return false, err, openai.ChatCompletionMessage{}
	}
	chanStream := make(chan string, 100)
	go func() {
		defer stream.Close()
		defer close(chanStream)
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println("Stream finished")
				chanStream <- "<!finish>"
				return
			}

			if err != nil {
				fmt.Printf("Stream error: %v\n", err)
				chanStream <- "<!error>"
				return
			}
			if len(response.Choices) == 0 {
				fmt.Println("Stream finished")
				chanStream <- "<!finish>"
				return
			}
			data := response.Choices[0].Delta.Content
			chanStream <- string(data)
			fmt.Printf("Stream response: %s\n", data)
		}
	}()
	var sendOk = true
	var aswMsg = ""
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-chanStream; ok {
			if msg == "<!finish>" {
				c.SSEvent("stop", "finish")
			}
			if msg == "<!error>" {
				c.SSEvent("stop", "error")
				sendOk = false
				msg += "请求失败,请重新提问"
			}
			c.SSEvent("message", msg)
			aswMsg = aswMsg + msg
			fmt.Printf("message: %v\n", msg)
			return true
		}
		return false
	})
	return sendOk, nil, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: aswMsg,
	}
}

package http

import (
	"context"
	// "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
	// "github.com/go-playground/locales/lg"
	"github.com/sashabaranov/go-openai"
	uuid "github.com/satori/go.uuid"
	"gpt3.5/cfg"
	"gpt3.5/gptcli"
	gptstream "gpt3.5/http/stream"
	gptws "gpt3.5/ws"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "website",
		})
	})

	r.GET("/ws", func(c *gin.Context) {
		ws, err := gptws.Upgrader.Upgrade(c.Writer, c.Request, nil)

		if err != nil {
			fmt.Println(err)
			return
		}

		go func() {
			var reqParam request
			err = c.ShouldBindQuery(&reqParam)
			if err != nil {
				log.Println(err.Error())
				return
			}

			fmt.Printf("reqpram is%#v \n", reqParam)

			if reqParam.Token == "" {
				//为了逐字显示
				msg := "未获取到token,请检查token设置!"
				for _, ruc := range msg {
					time.Sleep(100 * time.Nanosecond)
					ws.WriteMessage(websocket.TextMessage, []byte(string(ruc)))
				}
				if err = ws.Close(); err != nil {
					fmt.Printf("close ws_conn err:%v \n", err)
				}
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
				if ok, asw, err := WSProcess(ws, req); err != nil {
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
				if ok, asw, err := WSProcess(ws, req); err != nil {
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
		}()
		// go  gptws.HandleWs(ws)

	})

	r.POST("/v1/chat/do", func(c *gin.Context) {
		var reqParam request
		if err := c.ShouldBindJSON(&reqParam); err != nil {
			log.Println(err.Error())
			return
		}
		if reqParam.Token == "" {
			uuidTk := uuid.NewV4().String()
			apiParam := []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: cfg.Cfg.CharacterSetting,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: reqParam.Message,
				},
			}
			fmt.Printf("%+v", apiParam)
			apiRequest := openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: apiParam,
			}
			resp, err := gptcli.Cli().CreateChatCompletion(context.Background(), apiRequest)

			if err != nil {
				log.Printf("ChatCompletion error: %v\n", err)
				sErr := fmt.Sprintf("ChatCompletion error: %v", err)
				c.JSON(500, response{Err: sErr})
				return
			}
			apiParam = append(apiParam, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: resp.Choices[0].Message.Content,
			})
			gptcli.TokenManager.Store(uuidTk, &gptcli.Token{
				Context:  apiParam,
				LastTime: time.Now(),
			})
			httpResp := response{
				Asw:          resp.Choices[0].Message.Content,
				TokenRequire: uuidTk,
			}
			c.JSON(200, httpResp)
		} else {
			if v, exist := gptcli.TokenManager.Load(reqParam.Token); exist {
				t := v.(*gptcli.Token)
				// fmt.Printf("token is %+v", t)
				tCtx := append(t.Context, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: reqParam.Message,
				})
				apiRequest := openai.ChatCompletionRequest{
					Model:    openai.GPT3Dot5Turbo,
					Messages: tCtx,
				}
				resp, err := gptcli.Cli().CreateChatCompletion(context.Background(), apiRequest)
				if err != nil {
					log.Printf("ChatCompletion error: %v\n", err)
					sErr := fmt.Sprintf("ChatCompletion error: %v", err)
					c.JSON(500, response{Err: sErr, ErrCode: "500"})
					return
				}
				tCtx = append(tCtx, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: resp.Choices[0].Message.Content,
				})
				t.Context = tCtx
				//更新token时间
				t.LastTime = time.Now()
				httpResp := response{
					Asw: resp.Choices[0].Message.Content,
				}
				// fmt.Printf("%+v",resp)
				c.JSON(200, httpResp)
			} else {
				//token 不存在
				log.Println("invalid token")
				sErr := fmt.Sprintf("invalid token:\"%s\"", reqParam.Token)
				c.JSON(401, response{Err: sErr, ErrCode: "401"})
			}
		}
	})

	r.GET("/cfg/modkey", func(c *gin.Context) {
		auth := c.Request.Header.Get("x-auth")
		if auth == cfg.Cfg.SecretKey {
			apikey := c.Query("apikey")
			gptcli.SwitchCliWithApiKey(apikey)
			c.String(200, "ok")
			return
		}
		c.String(401, "invalid SecretKey")
	})

	r.GET("/stream", func(c *gin.Context) {
		var reqParam request
		err := c.ShouldBindQuery(&reqParam)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Printf("reqpram is%#v \n", reqParam)

		if reqParam.Token == "" {
			//为了逐字显示
			msg := []rune("未获取到token,请检查token设置!")
			idx := 0
			len := len(msg)
			c.Stream(func(w io.Writer) bool {
				time.Sleep(100 * time.Nanosecond)
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
			if ok, asw, err := SSEventProcess(c, req); err != nil {
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
			if ok, asw, err := SSEventProcess(c, req); err != nil {
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
func WSProcess(ws *websocket.Conn, req openai.ChatCompletionRequest) (bool, openai.ChatCompletionMessage, error) {
	stream, err := gptcli.Cli().CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		return false, openai.ChatCompletionMessage{}, err
	}
	streamChan := gptstream.StreamRequest(stream)
	var (
		sendOk = true
		aswMsg = ""
	)
	for msg := range streamChan {
		if msg == "<!finish>" {
			ws.Close()
			return sendOk, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: aswMsg,
			}, nil
		}
		if msg == "<!error>" {
			sendOk = false
			msg += "请求失败,请重新提问"
			ws.WriteMessage(websocket.TextMessage, []byte(msg))
			ws.Close()
			return sendOk, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: aswMsg,
			}, nil
		}
		ws.WriteMessage(websocket.TextMessage, []byte(msg))
		aswMsg = aswMsg + msg
		// fmt.Printf("message: %v\n", msg)
	}
	ws.Close()
	return sendOk, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: aswMsg,
	}, nil
}

func SSEventProcess(c *gin.Context, req openai.ChatCompletionRequest) (bool, openai.ChatCompletionMessage, error) {
	stream, err := gptcli.Cli().CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		return false, openai.ChatCompletionMessage{}, err
	}
	streamChan := gptstream.StreamRequest(stream)
	var (
		sendOk = true
		aswMsg = ""
	)
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-streamChan; ok {
			if msg == "<!finish>" {
				c.SSEvent("stop", "finish")
				fmt.Println(aswMsg)
				return false
			}
			if msg == "<!error>" {
				c.SSEvent("stop", "error")
				sendOk = false
				msg += "请求失败,请重新提问"
			}
			c.SSEvent("message", msg)
			aswMsg = aswMsg + msg
			// fmt.Printf("message: %v\n", msg)
			return true
		}
		return false
	})
	return sendOk, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: aswMsg,
	}, nil
}

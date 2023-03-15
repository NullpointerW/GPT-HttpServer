package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
	uuid "github.com/satori/go.uuid"
	"gpt3.5/gptcli"
)

type request struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type response struct {
	TokenRequire string `json:"tokenRequire"`
	Asw          string `json:"asw"`
}

func Do(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(405)
		_, _ = fmt.Fprintf(w, "%s", "non-post method not allowed")
		return
	}
	jsRaw, _ := io.ReadAll(req.Body)
	reqParam := &request{}
	err := json.Unmarshal(jsRaw, req)
	if err != nil {
		log.Println(err)
	}
	if reqParam.Token == "" {
		uuidTk := uuid.NewV4().String()
		apiParam := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "you are a teacher",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: reqParam.Message,
			},
		}
		apiRequest := openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: apiParam,
		}
		resp, err := gptcli.Cli.CreateChatCompletion(context.Background(), apiRequest)
		if err != nil {
			log.Printf("ChatCompletion error: %v\n", err)
			w.WriteHeader(500)
			_, _ = fmt.Fprintf(w, "ChatCompletion error: %v\n", err)
			return
		}
		gptcli.TokenManager.Store(uuidTk, &gptcli.Token{
			Context:  apiParam,
			LastTime: time.Now(),
		})
		httpresp := response{
			Asw:          resp.Choices[0].Message.Content,
			TokenRequire: uuidTk}
		jsonRaw, _ := json.Marshal(httpresp)
		w.WriteHeader(200)
		_, _ = fmt.Fprintf(w, "%s", jsonRaw)
	
	} else {
		if v, exist := gptcli.TokenManager.Load(reqParam.Token); exist {
			t := v.(*gptcli.Token)
			tctx := append(t.Context, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: reqParam.Message,
			})
			apiRequest := openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: tctx,
			}
			resp, err := gptcli.Cli.CreateChatCompletion(context.Background(), apiRequest)
			if err != nil {
				log.Printf("ChatCompletion error: %v\n", err)
				w.WriteHeader(500)
				_, _ = fmt.Fprintf(w, "ChatCompletion error: %v\n", err)
				return
			}
			t.Context = tctx
			//更新token时间
			t.LastTime=time.Now()
			httpresp := response{
				Asw: resp.Choices[0].Message.Content,
			}
			jsonRaw, _ := json.Marshal(httpresp)
			w.WriteHeader(200)
			_, _ = fmt.Fprintf(w, "%s", jsonRaw)
		} else {
			//token 不存在
			log.Println("invalid token")
			w.WriteHeader(500)
			_, _ = fmt.Fprintf(w, "invalid token \"%s\"",reqParam.Token )
		}
	}

}

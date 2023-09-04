package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/NullpointerW/go-openai"
	uuid "github.com/satori/go.uuid"
	"gpt3.5/cfg"
	"gpt3.5/gptcli"
)

type request struct {
	Token   string `json:"token" uri:"token" form:"token"`
	Message string `json:"message"  uri:"token" form:"message"`
}

type FineTunesRequest struct {
	Model   string `json:"model" uri:"model" form:"model"`
	Message string `json:"message"  uri:"token" form:"message"`
}

type response struct {
	TokenRequire string `json:"tokenRequire,omitempty"`
	Asw          string `json:"asw,omitempty"`
	Err          string `json:"error,omitempty"`
	ErrCode      string `json:"errCode,omitempty"`
}

// Deprecated:impl via gin
func Do(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(405)
		_, _ = fmt.Fprintf(w, "%s", "non-post method not allowed")
		return
	}
	jsRaw, _ := io.ReadAll(req.Body)
	fmt.Println(string(jsRaw))
	reqParam := &request{}
	err := json.Unmarshal(jsRaw, reqParam)
	if err != nil {
		log.Println(err)
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
			w.WriteHeader(500)
			sErr := fmt.Sprintf("ChatCompletion error: %v", err)
			jsonRaw, _ := json.Marshal(response{Err: sErr})
			_, _ = fmt.Fprintf(w, "%s", jsonRaw)
			return
		}
		apiParam = append(apiParam, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: resp.Choices[0].Message.Content,
		})
		gptcli.TokenManager.Store(uuidTk, &gptcli.PromptContext{
			Context:  apiParam,
			LastTime: time.Now(),
		})
		httpResp := response{
			Asw:          resp.Choices[0].Message.Content,
			TokenRequire: uuidTk,
		}
		// fmt.Printf("%+v",resp)
		jsonRaw, _ := json.Marshal(httpResp)
		w.WriteHeader(200)
		_, _ = fmt.Fprintf(w, "%s", jsonRaw)

	} else {
		if v, exist := gptcli.TokenManager.Load(reqParam.Token); exist {
			t := v.(*gptcli.PromptContext)
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
				w.WriteHeader(500)
				sErr := fmt.Sprintf("ChatCompletion error: %v", err)
				jsonRaw, _ := json.Marshal(response{Err: sErr, ErrCode: "500"})
				_, _ = fmt.Fprintf(w, "%s", jsonRaw)
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
			jsonRaw, _ := json.Marshal(httpResp)
			w.WriteHeader(200)
			_, _ = fmt.Fprintf(w, "%s", jsonRaw)
		} else {
			//token 不存在
			log.Println("invalid token")
			w.WriteHeader(401)
			sErr := fmt.Sprintf("invalid token:\"%s\"", reqParam.Token)
			jsonRaw, _ := json.Marshal(response{Err: sErr, ErrCode: "401"})
			_, _ = fmt.Fprintf(w, "%s", jsonRaw)
		}
	}
}

// Deprecated: impl via gin
func SwitchApikey(w http.ResponseWriter, req *http.Request) {
	auth := req.Header.Get("x-auth")
	if auth == cfg.Cfg.SecretKey {
		apikey := req.URL.Query().Get("apikey")
		gptcli.SwitchCliWithApiKey(apikey)
		w.WriteHeader(200)
		_, _ = fmt.Fprintf(w, "%s", "ok")
		return
	}
	w.WriteHeader(401)
	_, _ = fmt.Fprintf(w, "invalid SecretKey")
}

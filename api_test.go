package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/sashabaranov/go-openai"
	"gpt3.5/gptcli"
)

func TestApi(t *testing.T) {
	resp, err := gptcli.Cli.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "以下我的所有问题，你必须先核实问题是否与建筑施工质量安全领域相关，如果不相关，请不要回答",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Message.Content)
}

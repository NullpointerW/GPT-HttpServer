package main

import (
	"log"
	"net/http"
	gptHttp "gpt3.5/http"
)

func main() {

	mux := http.NewServeMux()
	mux.Handle("/chat/do", http.HandlerFunc(gptHttp.Do))
	log.Println("starting http-srv")
	log.Fatal(http.ListenAndServe("localhost:8000", mux))
	// resp, err := gptcli.Cli.CreateChatCompletion(
	// 	context.Background(),
	// 	openai.ChatCompletionRequest{
	// 		Model: openai.GPT3Dot5Turbo,
	// 		Messages: []openai.ChatCompletionMessage{
	// 			{
	// 				Role:    openai.ChatMessageRoleSystem,
	// 				Content: "you are a teacher",
	// 			},
	// 			{
	// 				Role:    openai.ChatMessageRoleUser,
	// 				Content: "Hello!",
	// 			},
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	fmt.Printf("ChatCompletion error: %v\n", err)
	// 	return
	// }

	// fmt.Println(resp.Choices[0].Message.Content)

	// client := openai.NewClient("your token")
	// resp, err := client.CreateChatCompletion(
	// 	context.Background(),
	// 	openai.ChatCompletionRequest{
	// 		Model: openai.GPT3Dot5Turbo,
	// 		Messages: []openai.ChatCompletionMessage{
	// 			{
	// 				Role:    openai.ChatMessageRoleUser,
	// 				Content: "Hello!",
	// 			},
	// 		},
	// 	},
	// )

	// if err != nil {
	// 	fmt.Printf("ChatCompletion error: %v\n", err)
	// 	return
	// }

	// fmt.Println(resp.Choices[0].Message.Content)

}

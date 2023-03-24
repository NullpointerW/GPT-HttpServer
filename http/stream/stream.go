package stream

import (
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

func StreamRequest(st *openai.ChatCompletionStream) chan string {
	ch := make(chan string, 100)
	go process(st, ch)
	return ch
}

func process(stream *openai.ChatCompletionStream, streamCh chan string) {
	defer stream.Close()
	defer close(streamCh)
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			streamCh <- "<!finish>"
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			streamCh <- "<!error>"
			return
		}
		if len(response.Choices) == 0 {
			fmt.Println("Stream finished")
			streamCh <- "<!finish>"
			return
		}
		data := response.Choices[0].Delta.Content
		streamCh <- string(data)
		fmt.Printf("Stream response: %s\n", data)
	}
}

package fine_tunes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/NullpointerW/go-openai"
	uuid "github.com/satori/go.uuid"
	"gpt3.5/gptcli"
	"io"
	"os"
	"time"
)

type QA struct {
	Q string `json:"q"`
	A string `json:"a"`
}

func (qa QA) BuildJson(w io.Writer) error {
	type Msg []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	prompt := struct {
		Message Msg `json:"message"`
	}{Msg{{"user", qa.Q}, {"assistant", qa.A}}}
	err := json.NewEncoder(w).Encode(prompt)
	if err != nil {
		return err
	}
	//_, err = w.Write([]byte("\n"))
	//if err != nil {
	//	return err
	//}
	return nil
}

func CreateFineTune(qas []QA) error {
	fn := uuid.NewV4().String() + ".jsonl"
	file, err := os.OpenFile(fn, os.O_TRUNC|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	for _, qa := range qas {
		err := qa.BuildJson(file)
		if err != nil {
			return err
		}
	}
	go FineTuneProcess(file)
	return nil
}

func FineTuneProcess(file *os.File) {
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
		err = os.Remove(file.Name())
		if err != nil {
			fmt.Println(err)
		}
	}()
	cli := gptcli.Cli()
	openaiFile, err := cli.CreateFile(context.Background(), openai.FileRequest{
		FilePath: file.Name(),
		Purpose:  "fine-tune",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()
	if openaiFile.Status != "processed" {
		for {
			openaiFile, _ = cli.GetFile(ctx, openaiFile.ID)
			if openaiFile.Status == "processed" {
				break
			}
			time.Sleep(time.Millisecond * 500)
		}
	}
	fineTuningJob, err := cli.CreateFineTuningJob(ctx, openai.FineTuningJobRequest{
		TrainingFile: openaiFile.ID,
		Model:        "gpt-3.5-turbo", // gpt-3.5-turbo-0613, babbage-002.
	})
	if err != nil {
		fmt.Printf("Creating new fine tune model error: %v\n", err)
		return
	}
	if fineTuningJob.Status != "succeeded" {
		for {
			fineTuningJob, err = cli.RetrieveFineTuningJob(ctx, fineTuningJob.ID)
			if err != nil {
				fmt.Printf("Getting fine tune model error: %v\n", err)
				return
			}
			if fineTuningJob.Status == "succeeded" {
				break
			}
			time.Sleep(5 * time.Minute)
		}
	}

	fineTuningJob, err = cli.RetrieveFineTuningJob(ctx, fineTuningJob.ID)
	if err != nil {
		fmt.Printf("Getting fine tune model error: %v\n", err)
		return
	}

}

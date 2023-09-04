package fine_tunes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/NullpointerW/go-openai"
	uuid "github.com/satori/go.uuid"
	"gpt3.5/cache"
	"gpt3.5/gptcli"
	"io"
	"os"
	"strings"
	"time"
)

type QA struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type FineTuneModel struct {
	Name           string `json:"name"`
	OpenaiFileId   string `json:"OpenaiFileId"` // random gen uuid
	FineTuneJobId  string `json:"fineTuneJobId"`
	FileId         string `json:"fileId"`
	Done           bool   `json:"done"`
	Model          string `json:"model"`
	FileUpLoadDone bool   `json:"fileUpLoadDone"`
}

func (qa QA) BuildJson(w io.Writer) error {
	type Msg []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	prompt := struct {
		Message Msg `json:"messages"`
	}{Msg{{"user", qa.Q}, {"assistant", qa.A}}}
	err := json.NewEncoder(w).Encode(prompt)
	if err != nil {
		return err
	}
	return nil
}

func CreateFineTune(qas []QA, uid, name string) error {
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
	go FineTuneProcess(file, uid, name, false, "")
	select {}
	return nil
}

func FineTuneProcess(file *os.File, uid, name string, reload bool, reLocFTId string) {
	var (
		ctx                = context.Background()
		cli                = gptcli.Cli()
		fineTuningJob      = openai.FineTuningJob{}
		marshal            []byte
		localizeFineTuneId string
		ftMode             = FineTuneModel{}
		openaiFile         = openai.File{}
		err                error
	)
	if reload {
		localizeFineTuneId = reLocFTId
		get := cache.Redis.HGet(uid, localizeFineTuneId)
		marshal, err = get.Bytes()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(marshal, &ftMode)
		if err != nil {
			fmt.Println(err)
			return
		}
		if !ftMode.FileUpLoadDone {
			goto uploadFile
		}
		if !ftMode.Done {
			goto fineTuneTask
		}

	}
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

	openaiFile, err = cli.CreateFile(context.Background(), openai.FileRequest{
		FilePath: file.Name(),
		Purpose:  "fine-tune",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	localizeFineTuneId = strings.TrimSuffix(file.Name(), ".jsonl")

	ftMode = FineTuneModel{OpenaiFileId: openaiFile.ID,
		FileUpLoadDone: openaiFile.Status == "processed",
		Name:           name}
	marshal, _ = json.Marshal(ftMode)
	cache.Redis.HSet(uid, localizeFineTuneId, marshal)
uploadFile:
	if openaiFile.Status != "processed" {
		for {
			openaiFile, _ = cli.GetFile(ctx, ftMode.OpenaiFileId)
			if openaiFile.Status == "processed" {
				ftMode.FileUpLoadDone = true
				marshal, err := json.Marshal(ftMode)
				if err != nil {
					fmt.Println(err)
					return
				}
				cache.Redis.HSet(uid, localizeFineTuneId, marshal)
				fmt.Println("ft-model-id(local)", localizeFineTuneId, "openai-file-id", openaiFile.ID, "upload ok!")
				break
			}
			fmt.Println("ft-model-id(local)", localizeFineTuneId, "openai-file-id", openaiFile.ID, "not upload yet!")
			time.Sleep(time.Millisecond * 500)
		}
	}

	fineTuningJob, err = cli.CreateFineTuningJob(ctx, openai.FineTuningJobRequest{
		TrainingFile: openaiFile.ID,
		Model:        "gpt-3.5-turbo", // gpt-3.5-turbo-0613, babbage-002.
	})
	if err != nil {
		fmt.Printf("Creating new fine tune model error: %v\n", err)
		return
	}
	ftMode.FineTuneJobId = fineTuningJob.ID
	ftMode.Done = fineTuningJob.Status == "succeeded"
	ftMode.Model = fineTuningJob.FineTunedModel
	marshal, err = json.Marshal(ftMode)
	if err != nil {
		fmt.Println(err)
		return
	}
	cache.Redis.HSet(uid, localizeFineTuneId, marshal)
fineTuneTask:
	if fineTuningJob.Status != "succeeded" {
		for {
			fineTuningJob, err = cli.RetrieveFineTuningJob(ctx, fineTuningJob.ID)
			if err != nil {
				fmt.Printf("Getting fine tune model error: %v\n", err)
				return
			}
			if fineTuningJob.Status == "succeeded" {
				ftMode.Done = true
				ftMode.Model = fineTuningJob.FineTunedModel
				marshal, err := json.Marshal(ftMode)
				if err != nil {
					fmt.Println(err)
					return
				}
				cache.Redis.HSet(uid, localizeFineTuneId, marshal)
				fmt.Println("fineTuneJob is done ,goroutine exited")
				break
			}
			fmt.Println("ft-model-id(local)", localizeFineTuneId, "openai-finetunes-id", ftMode.FineTuneJobId, "not fin yet!", "status", fineTuningJob.Status)
			time.Sleep(2 * time.Minute)
		}
	}

}

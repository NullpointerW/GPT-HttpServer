package fine_tunes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/NullpointerW/go-openai"
	uuid "github.com/satori/go.uuid"
	"gpt-http/cache"
	"gpt-http/gptcli"
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

func init() {
	for _, k := range cache.Keys() {
		hAll, err := cache.HGetAll(k)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for id, _ := range hAll {
			goroutineId := context.WithValue(context.Background(), "id", k+"::"+id)
			go FineTuneProcess(nil, k, "", true, id, goroutineId)
		}
	}
}

func FinTuneList(uid string) (json.RawMessage, error) {
	var jsonRaws []json.RawMessage
	hAll, err := cache.HGetAll(uid)
	if err != nil {
		return []byte(""), err
	}
	for _, v := range hAll {
		jsonRaws = append(jsonRaws, json.RawMessage(v))
	}
	marshal, err := json.Marshal(jsonRaws)
	if err != nil {
		return []byte(""), err
	}
	return marshal, nil

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
	goroutineId := context.WithValue(context.Background(), "id", "create-goroutine::"+uid)
	go FineTuneProcess(file, uid, name, false, "", goroutineId)
	//select {}
	return nil
}

func FineTuneProcess(file *os.File, uid, name string, reload bool, reLocFTId string, goroutineId context.Context) {
	var (
		ctx                = context.Background()
		cli                = gptcli.Cli()
		fineTuningJob      = openai.FineTuningJob{}
		localizeFineTuneId string
		ftMode             = FineTuneModel{}
		openaiFile         = openai.File{}
		err                error
	)
	if reload {
		localizeFineTuneId = reLocFTId
		err := cache.HGet(uid, localizeFineTuneId, &ftMode)
		if err != nil {
			fmt.Println(err)
			return
		}
		if !ftMode.FileUpLoadDone {
			goto uploadFile
		}
		if !ftMode.Done {
			fineTuningJob.ID = ftMode.FineTuneJobId
			goto fineTuneTask
		}
		return
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
	_ = cache.HSet(uid, localizeFineTuneId, ftMode)
uploadFile:
	if openaiFile.Status != "processed" {
		c, repeat := 0, 20
		for {
			openaiFile, _ = cli.GetFile(ctx, ftMode.OpenaiFileId)
			if openaiFile.Status == "processed" {
				ftMode.FileUpLoadDone = true
				_ = cache.HSet(uid, localizeFineTuneId, ftMode)
				fmt.Println("ft-model-id(local)", localizeFineTuneId, "openai-file-id", openaiFile.ID, "upload ok!")
				break
			}
			c++
			if c >= 20 {
				fmt.Println("ft-model-id(local)", localizeFineTuneId, "request time limited,repeat:", repeat)
				return
			}
			fmt.Println("ft-model-id(local)", localizeFineTuneId, "openai-file-id", openaiFile.ID, "not upload yet!", "call times", c)
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
	_ = cache.HSet(uid, localizeFineTuneId, ftMode)
fineTuneTask:
	if fineTuningJob.Status != "succeeded" {
		c, repeat := 0, 20
		for {
			fineTuningJob, err = cli.RetrieveFineTuningJob(ctx, fineTuningJob.ID)
			if err != nil {
				fmt.Printf("%s:Getting fine tune model error: %v\n", goroutineId.Value("id"), err)
				return
			}
			if fineTuningJob.Status == "succeeded" {
				ftMode.Done = true
				ftMode.Model = fineTuningJob.FineTunedModel
				_ = cache.HSet(uid, localizeFineTuneId, ftMode)
				fmt.Println("fineTuneJob is done ,goroutine exited")
				break
			}
			c++
			if c >= 20 {
				fmt.Println("ft-model-id(local)", localizeFineTuneId, "request time limited,repeat:", repeat)
				return
			}
			fmt.Println("ft-model-id(local)", localizeFineTuneId, "openai-finetunes-id", ftMode.FineTuneJobId, "not fin yet!", "status", fineTuningJob.Status,
				"call times", c)
			time.Sleep(2 * time.Minute)
		}
	}
}

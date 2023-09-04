package gptcli

import (
	"fmt"
	"time"
)

func tokensCleaner(d time.Duration) {
sleep:
	for {
		time.Sleep(d)
		goto clean
	}
clean:
	its := func(k, v interface{}) bool {
		tk := v.(*PromptContext)
		if time.Since(tk.LastTime) >= d {
			TokenManager.Delete(k)
			fmt.Printf("clean token %s \n", k.(string))
		}
		return true
	}
	TokenManager.Range(its)
	goto sleep
}

func fineTunesCleaner(d time.Duration) {
sleep:
	for {
		time.Sleep(d)
		goto clean
	}
clean:
	its := func(k, v any) bool {
		tk := v.(*PromptContext)
		if time.Since(tk.LastTime) >= d || len(tk.Context) >= 2000 {
			FineTunesManager.Delete(k)
			fmt.Printf("clean fineTunes prompts %s \n", k.(string))
		}
		return true
	}
	FineTunesManager.Range(its)
	goto sleep
}

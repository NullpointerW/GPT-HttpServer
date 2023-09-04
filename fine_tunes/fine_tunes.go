package fine_tunes

import (
	"encoding/json"
	"io"
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

func CreateFineTune() {

}

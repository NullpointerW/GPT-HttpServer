package fine_tunes

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestQA_BuildJson(t *testing.T) {
	type fields struct {
		Q string
		A string
	}
	tests := []struct {
		name    string
		fields  fields
		wantW   string
		wantErr bool
	}{{"t1", fields{"hello", "hi"}, "w", false}} // TODO: Add test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qa := QA{
				Q: tt.fields.Q,
				A: tt.fields.A,
			}
			w := &bytes.Buffer{}
			err := qa.BuildJson(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("BuildJson() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestQA_BuildJson2(t *testing.T) {
	file, err := os.OpenFile("./mytest.jsonl", os.O_TRUNC|os.O_CREATE, 0777)
	if err != nil {
		t.Error(err)
		return
	}

	qas := []QA{{"he", "w"}, {"he", "w"}, {"he", "w"}}
	for _, qa := range qas {
		err := qa.BuildJson(file)
		if err != nil {
			t.Error(err)
			return
		}
	}

}

func TestCreateFineTune(t *testing.T) {
	qas := []QA{{"chong", "wong"}, {"he", "w"}, {"he", "wwa12w"}, {"w", "y"}, {"he", "w"}, {"he", "w"}, {"he", "w"}, {"he", "w"}, {"he", "w"}, {"he", "w"}, {"he", "w"}}
	CreateFineTune(qas, "wkc", "mytest-model")
}

func TestFinTuneList(t *testing.T) {
	list, err := FinTuneList("wkc")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(list)

}

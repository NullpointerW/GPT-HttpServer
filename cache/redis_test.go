package cache

import (
	"fmt"
	"testing"
)

func TestHSet(t *testing.T) {
	type args struct {
		Name string
		Test string
	}
	a := args{"myfuckingtest", "hell"}
	if err := HSet("test", "hkey", a); err != nil {
		t.Error(err)
		return
	}
}

func TestHGet(t *testing.T) {
	type args struct {
		Name string
		Test string
	}
	var a args
	if err := HGet("test", "hkey", &a); err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%+v ?", a)
}

func TestKeys(t *testing.T) {
	fmt.Println(Keys())
}

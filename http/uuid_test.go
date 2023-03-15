package http

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestUUid(t *testing.T) {
	fmt.Println(uuid.NewV4().String())
}

package test

import (
	"github.com/nickbadlose/muzz/router"
	"net/http/httptest"
	"testing"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

// TODO table tests?

func newRequest() {}

func TestStatus(t *testing.T) {
	srv := httptest.NewServer(router.New())
	defer srv.Close()
	t.Run("/status", func(t *testing.T) {
	})
}

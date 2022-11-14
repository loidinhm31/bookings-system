package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var testH testHandler
	h := NoSurf(&testH)

	switch v := h.(type) {
	case http.Handler:
	// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var testH testHandler
	h := SessionLoad(&testH)

	switch v := h.(type) {
	case http.Handler:
	// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}

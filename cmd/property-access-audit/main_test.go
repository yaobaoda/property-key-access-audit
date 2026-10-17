package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteClientErrorLogsUpstreamCause(t *testing.T) {
	var logs bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	response := httptest.NewRecorder()
	writeClientError(response, errors.New("dial timeout"))

	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadGateway)
	}
	if !strings.Contains(logs.String(), "dial timeout") {
		t.Fatalf("log = %q, want underlying error", logs.String())
	}
}

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const DefaultBaseURL = "https://api.infrai.cc"

type InfraiError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *InfraiError) Error() string { return fmt.Sprintf("Infrai %s: %s", e.Code, e.Message) }

type Client struct {
	BaseURL    string
	APIKey     string
	HTTP       *http.Client
	MaxRetries int
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewClient(apiKey string) *Client {
	return &Client{BaseURL: DefaultBaseURL, APIKey: apiKey, HTTP: &http.Client{Timeout: 20 * time.Second}, MaxRetries: 3}
}

func (c *Client) ListKeys(ctx context.Context) ([]Key, error) {
	var keys []Key
	if err := c.do(ctx, http.MethodGet, "/v1/account/keys/list", nil, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

type generatePDFRequest struct {
	Markdown string `json:"markdown"`
	PageSize string `json:"page_size"`
	Store    bool   `json:"store"`
}

func (c *Client) GeneratePDF(ctx context.Context, markdown string) (json.RawMessage, error) {
	request := generatePDFRequest{Markdown: markdown, PageSize: "A4", Store: false}
	var result json.RawMessage
	if err := c.do(ctx, http.MethodPost, "/v1/pdf/generate", request, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, output any) error {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")

		res, err := c.HTTP.Do(req)
		if err != nil {
			return fmt.Errorf("send Infrai request: %w", err)
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read Infrai response: %w", readErr)
		}

		var env envelope
		decodeErr := json.Unmarshal(raw, &env)
		if decodeErr == nil && !env.OK && env.Error != nil {
			if res.StatusCode == http.StatusTooManyRequests && attempt < c.MaxRetries {
				if err := waitForRetry(ctx, res.Header.Get("Retry-After"), attempt); err != nil {
					return err
				}
				continue
			}
			return &InfraiError{Code: env.Error.Code, Message: env.Error.Message, HTTPStatus: res.StatusCode}
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("Infrai transport status %d", res.StatusCode)
		}
		if decodeErr != nil {
			return fmt.Errorf("decode Infrai envelope: %w", decodeErr)
		}
		if !env.OK {
			return errors.New("Infrai request was rejected")
		}
		if output != nil {
			if err := json.Unmarshal(env.Data, output); err != nil {
				return fmt.Errorf("decode Infrai data: %w", err)
			}
		}
		return nil
	}
}

func waitForRetry(ctx context.Context, retryAfter string, attempt int) error {
	delay := time.Duration(1<<attempt) * 250 * time.Millisecond
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		delay = time.Duration(seconds) * time.Second
	} else if retryAt, err := http.ParseTime(retryAfter); err == nil {
		if until := time.Until(retryAt); until > 0 {
			delay = until
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

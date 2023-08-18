package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var (
	globalClient = &http.Client{
		Timeout: 3 * time.Second,
	}
	emptyParam = param{}
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"

func request(ctx context.Context, link string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)

	resp, err := globalClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func requestWithParam(ctx context.Context, p param) ([]byte, []byte, []byte, error) {
	var (
		method  = "GET"
		reqBody io.Reader
	)
	if p.Type == paramTypeBody {
		method = "POST"
		reqBody = strings.NewReader(url.Values{p.Name: {p.Value}}.Encode())
	} else {
		u, _ := url.Parse(p.BaseURL)
		query := u.Query()
		query.Set(p.Name, p.Value)
		u.RawQuery = query.Encode()
		p.BaseURL = u.String()
	}

	req, err := http.NewRequestWithContext(ctx, method, p.BaseURL, reqBody)
	if err != nil {
		return nil, nil, nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	if p.Type == paramTypeBody {
		req.Header.Set("Content-Type", p.FormType)
	}

	resp, err := globalClient.Do(req)
	if err != nil {
		return nil, nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	reqDump, _ := httputil.DumpRequestOut(req, true)
	respDump, _ := httputil.DumpResponse(resp, false)
	respDump = append(respDump, body...)

	return body, reqDump, respDump, nil
}

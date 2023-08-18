package main

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

const (
	paramTypeQuery = "query"
	paramTypePath  = "path"
	paramTypeBody  = "body"
)

type param struct {
	Name     string `json:"name"`
	Token    string `json:"-"`
	Value    string `json:"-"`
	BaseURL  string `json:"base_url"`
	Type     string `json:"type"`
	FormType string `json:"form_type"` // TODO: set content type for body type param
}

// collectParam find params
func collectParam(ctx context.Context, target url.URL) []param {
	logrus.Info("collecting params")

	resChan := make(chan param, 10)
	resChanDone := make(chan struct{})
	params := make([]param, 0)

	go func() {
		defer func(done chan<- struct{}) {
			done <- struct{}{}
		}(resChanDone)

		existMark := make(map[string]struct{})
		for {
			select {
			case <-ctx.Done():
				return
			case p, ok := <-resChan:
				if !ok {
					return
				}
				markLabel := p.Type + "/" + p.Name
				if _, ok := existMark[markLabel]; ok {
					continue
				}

				existMark[markLabel] = struct{}{}
				p.Token = randString(5)
				if p.BaseURL == "" {
					p.BaseURL = target.String()
				}
				params = append(params, p)
			}
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go getParamFromURL(ctx, &wg, target, resChan)
	go getParamFromHTML(ctx, &wg, target, resChan)
	// go getParamFromWordlist(ctx, &wg, target, resChan)
	wg.Wait()
	close(resChan)
	<-resChanDone

	return params
}

func getParamFromURL(ctx context.Context, wg *sync.WaitGroup, target url.URL, resChan chan<- param) {
	defer wg.Done()

	// TODO: path param

	// query param
	urlParam := target.Query()
	for name := range urlParam {
		resChan <- param{
			Name:    name,
			Type:    paramTypeQuery,
			BaseURL: target.String(),
		}
	}
}

func getParamFromHTML(ctx context.Context, wg *sync.WaitGroup, target url.URL, resChan chan<- param) {
	defer wg.Done()

	body, err := request(ctx, target.String())
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	wgUrl := sync.WaitGroup{}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && href != "" {
			u, err := url.Parse(href)
			if err != nil {
				return
			}
			wgUrl.Add(1)
			go getParamFromURL(ctx, &wgUrl, *u, resChan)
		}
	})
	doc.Find("form").Each(func(i int, s *goquery.Selection) { // TODO: add example test
		var (
			pType     = paramTypeBody
			pBaseURL  = target
			pFormType = "application/x-www-form-urlencoded"
		)
		formMethod, _ := s.Attr("method")
		if strings.ToLower(formMethod) == "get" {
			pType = paramTypeQuery
			pFormType = ""
		}
		formAction, _ := s.Attr("action")
		if formAction != "" {
			u, err := url.Parse(formAction)
			if err != nil {
				return
			}
			pBaseURL = *target.ResolveReference(u)
			wgUrl.Add(1)
			go getParamFromURL(ctx, &wgUrl, pBaseURL, resChan)
		}

		s.Find("input,select,textarea").Each(func(i int, s *goquery.Selection) {
			name, exists := s.Attr("name")
			if exists && name != "" {
				resChan <- param{
					Name:     name,
					Type:     pType,
					BaseURL:  pBaseURL.String(),
					FormType: pFormType,
				}
			}
		})

	})

	wgUrl.Wait()
}

// https://github.com/lutfumertceylan/top25-parameter
// https://paper.seebug.org/1119/#_1
var topDict = []string{"redirect", "redir", "url", "link", "goto", "debug", "_debug", "test",
	"get", "index", "src", "source", "file", "frame", "config", "new", "old", "var", "rurl", "return_to",
	"_return", "returl", "last", "text", "load", "email", "mail", "user", "username", "password", "pass",
	"passwd", "first_name", "last_name", "back", "href", "ref", "data", "input", "out", "net", "host",
	"address", "code", "auth", "userid", "auth_token", "token", "error", "keyword", "key", "q", "query",
	"aid", "bid", "cid", "did", "eid", "fid", "gid", "hid", "iid", "jid", "kid", "lid", "mid", "nid", "oid",
	"pid", "qid", "rid", "sid", "tid", "uid", "vid", "wid", "xid", "yid", "zid", "cal", "country", "x", "y",
	"topic", "title", "head", "higher", "lower", "width", "height", "add", "result", "log", "demo", "example",
	"message", "s", "search", "id", "lang", "page", "keywords", "year", "view", "type", "name", "p", "month",
	"image", "list_type", "terms", "categoryid", "login", "begindate", "enddate",
}

func getParamFromWordlist(ctx context.Context, wg *sync.WaitGroup, target url.URL, recChan chan<- param) {
	defer wg.Done()

	for _, name := range topDict {
		recChan <- param{
			Name:    name,
			Type:    paramTypeQuery,
			BaseURL: target.String(),
		}
		recChan <- param{
			Name:     name,
			Type:     paramTypeBody,
			BaseURL:  target.String(),
			FormType: "application/x-www-form-urlencoded",
		}
	}
}

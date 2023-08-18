package main

import (
	"bytes"
	"context"
	"net/url"

	"github.com/sirupsen/logrus"
)

func scan(ctx context.Context, target *url.URL) ([]*Vuln, error) {
	params := collectParam(ctx, *target)
	logrus.Infof("found %d params\n", len(params))

	res := make([]*Vuln, 0)
	for _, p := range params {
		vulns, err := (&scanWorker{}).Do(ctx, p)
		if err != nil {
			logrus.WithField("param", p.Name).Error(err)
			continue
		}
		if len(vulns) != 0 {
			res = append(res, vulns...)
		}
	}

	return res, nil
}

type scanWorker struct {
	rawHTML    []byte
	param      param
	tokenCount int
	logger     *logrus.Entry
}

func (s *scanWorker) Do(ctx context.Context, p param) ([]*Vuln, error) {
	s.logger = logrus.WithField("param", p.Name)

	p.Value = p.Token
	body, _, _, err := requestWithParam(ctx, p)
	if err != nil {
		return nil, err
	}
	tokenCount := bytes.Count(body, []byte(p.Token))
	if tokenCount == 0 {
		return nil, nil
	}
	s.tokenCount = tokenCount
	s.rawHTML = body
	s.param = p

	res := make([]*Vuln, 0, tokenCount)
	htmlVuln, err := s.scanHTML(ctx)
	if err != nil {
		s.logger.Error(err)
	}
	res = append(res, htmlVuln...)
	return res, nil
}

type payloadBuilder struct {
	prefix, suffix string
	Attempt        []string
}

func newPayloadBuilder(prefix, suffix string) *payloadBuilder {
	return &payloadBuilder{
		prefix:  prefix,
		suffix:  suffix,
		Attempt: make([]string, 0),
	}
}

func (p *payloadBuilder) Build(s string) string {
	pp := p.prefix + s + p.suffix
	p.Attempt = append(p.Attempt, pp)
	return pp
}

func (p *payloadBuilder) ResetXfix(pre, suf string) {
	p.prefix, p.suffix = pre, suf
}

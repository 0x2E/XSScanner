package main

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func (s *scanWorker) scanHTML(ctx context.Context) ([]*Vuln, error) {
	logger := s.logger.WithField("module", "html")

	// note: html lib will normalize original html,
	// 1. change tag name to lowercase
	// 2. change attr name to lowercase TODO: confirm
	doc, err := html.Parse(bytes.NewReader(s.rawHTML))
	if err != nil {
		return nil, err
	}

	res := make([]*Vuln, 0, s.tokenCount)

	var dfs func(*html.Node)
	dfs = func(n *html.Node) {
		if s.tokenCount < 1 {
			return
		}

		switch n.Type {
		case html.TextNode, html.CommentNode:
			tokenFound := strings.Count(n.Data, s.param.Token)
			if tokenFound > 0 {
				vuln, err := handleHTMLText(ctx, n, s.param)
				if err != nil {
					logger.WithField("type", "html-text&comment").Error(err)
				}
				s.tokenCount -= tokenFound
				res = append(res, vuln)
			}
		case html.ElementNode:
			// element name
			tokenFound := strings.Count(n.Data, strings.ToLower(s.param.Token)) // the parsing lib lowers element name
			if tokenFound > 0 {
				vuln, err := handleHTMLTagName(ctx, n, s.param)
				if err != nil {
					logger.WithField("type", "html-tagname").Error(err)
				}
				s.tokenCount -= tokenFound
				res = append(res, vuln)
			}

			// attr
			tokenFound = 0
			for _, attr := range n.Attr {
				tokenFound += strings.Count(attr.Val, s.param.Token)
				tokenFound += strings.Count(attr.Key, s.param.Token)
			}
			if tokenFound > 0 {
				vuln, err := handleHTMLAttr(ctx, n, s.param, s.rawHTML)
				if err != nil {
					logger.WithField("type", "html-attr").Error(err)
				}
				s.tokenCount -= tokenFound
				res = append(res, vuln...)
			}
		default:
		}

		if s.tokenCount > 0 {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				dfs(c)
			}
		}
	}

	dfs(doc)
	return res, err
}

func handleHTMLText(ctx context.Context, n *html.Node, p param) (*Vuln, error) {
	payload := newPayloadBuilder("", "")
	if n.Type == html.TextNode {
		parentTag := n.Parent.Data
		if parentTag != "html" {
			parentTag = randCap(parentTag)
			payload.ResetXfix("</"+parentTag+">", "<"+parentTag+">")
		}
	} else if n.Type == html.CommentNode {
		payload.ResetXfix("-->", "<!--")
	}

	return injHTMLElement(ctx, p, payload)
}

func handleHTMLTagName(ctx context.Context, n *html.Node, p param) (*Vuln, error) {
	payload := newPayloadBuilder(">", "<"+randCap(n.Data))

	return injHTMLElement(ctx, p, payload)
}

// https://portswigger.net/web-security/cross-site-scripting/cheat-sheet
var htmlXSSAttr = []string{
	"style", "href", "src",

	// event handler
	// https://www.w3schools.com/tags/ref_eventattributes.asp
	"onafterprint", "onafterscriptexecute", "onanimationcancel", "onanimationend", "onanimationiteration",
	"onanimationstart", "onauxclick", "onbeforecopy", "onbeforecut", "onbeforeinput", "onbeforeprint",
	"onbeforescriptexecute", "onbeforeunload", "onbegin", "onblur", "onbounce", "oncanplay", "oncanplaythrough",
	"onchange", "onclick", "onclose", "oncontextmenu", "oncopy", "oncuechange", "oncut", "ondblclick", "ondrag",
	"ondragend", "ondragenter", "ondragleave", "ondragover", "ondragstart", "ondrop", "ondurationchange", "onend",
	"onended", "onerror", "onfinish", "onfocus", "onfocusin", "onfocusout", "onfullscreenchange", "onhashchange",
	"oninput", "oninvalid", "onkeydown", "onkeypress", "onkeyup", "onload", "onloadeddata", "onloadedmetadata",
	"onmessage", "onmousedown", "onmouseenter", "onmouseleave", "onmousemove", "onmouseout", "onmouseover",
	"onmouseup", "onmousewheel", "onmozfullscreenchange", "onpagehide", "onpageshow", "onpaste", "onpause",
	"onplay", "onplaying", "onpointerdown", "onpointerenter", "onpointerleave", "onpointermove", "onpointerout",
	"onpointerover", "onpointerrawupdate", "onpointerup", "onpopstate", "onprogress", "onratechange", "onrepeat",
	"onreset", "onresize", "onscroll", "onscrollend", "onsearch", "onseeked", "onseeking", "onselect",
	"onselectionchange", "onselectstart", "onshow", "onstart", "onsubmit", "ontimeupdate", "ontoggle", "ontouchend",
	"ontouchmove", "ontouchstart", "ontransitioncancel", "ontransitionend", "ontransitionrun", "ontransitionstart",
	"onunhandledrejection", "onunload", "onvolumechange", "onwebkitanimationend", "onwebkitanimationiteration",
	"onwebkitanimationstart", "onwebkittransitionend", "onwheel",
}

func handleHTMLAttr(ctx context.Context, n *html.Node, p param, rawHTML []byte) ([]*Vuln, error) {
	logger := logrus.WithFields(logrus.Fields{
		"param": p.Name,
		"type":  "html-attr",
	})

	var (
		res     = make([]*Vuln, 0)
		badAttr = make([]html.Attribute, 0)
	)

	for _, attr := range n.Attr {
		bad := false
		if strings.Contains(attr.Key, p.Token) {
			bad = true
		}
		if strings.Contains(attr.Val, p.Token) {
			bad = true
			// inj in dangerous attr
			for _, v := range htmlXSSAttr {
				if attr.Key == v {
					res = append(res, &Vuln{
						Title: "dangerous attr: " + attr.Key,
						Param: p,
					})
					break
				}
			}
		}
		if bad {
			badAttr = append(badAttr, attr)
		}
	}

	// try escape
	for _, attr := range badAttr {
		payload := newPayloadBuilder("", "")
		// if injection position is in attr val
		if strings.Contains(attr.Val, p.Token) {
			// get the symbol wrapped around attr value,
			// the symbol may be " or ' or none
			re, err := regexp.Compile(fmt.Sprintf(`%s=(["']?)%s(["']?)`, attr.Key, p.Token))
			if err != nil {
				return nil, err
			}
			match := re.FindStringSubmatch(string(rawHTML))
			if len(match) < 3 {
				continue
			}
			wrappedSymbol := match[1:]
			payload.ResetXfix(wrappedSymbol[0]+">", wrappedSymbol[1]+"<x")
		}
		vuln, err := injHTMLElement(ctx, p, payload)
		if err != nil {
			logger.Error(err)
		}
		if vuln != nil {
			res = append(res, vuln)
		}
	}
	return res, nil
}

var tagFuzz = map[string]string{
	"img":    "imgsrconerror",
	"svg":    "svgonload",
	"iframe": "iframesrcjavascript",
	"a":      "ahrefjavascript",
	"input":  "inputautofocusonfocus",
}

func injHTMLElement(ctx context.Context, p param, payload *payloadBuilder) (*Vuln, error) {
	logger := logrus.WithFields(logrus.Fields{
		"param": p.Name,
		"type":  "html-element-inj",
	})
	// test <tag/>
	p.Value = payload.Build("<" + p.Token + "/>")
	body, _, _, err := requestWithParam(ctx, p)
	if err != nil {
		return nil, err
	}

	logger.Debug("payload: [" + p.Value + "] resp: " + string(body))
	if !containsHTMLElement(body, p.Token) {
		return nil, nil
	}

	var vuln *Vuln
	// fuzz tag
	for _, s := range tagFuzz {
		tp := randCap(s + randString(3))
		p.Value = payload.Build(tp)
		body, reqDump, respDump, err := requestWithParam(ctx, p)
		if err != nil {
			continue
		}
		logger.Debug("payload: [" + p.Value + "] resp: " + string(body))
		if bytes.Contains(body, []byte(tp)) {
			vuln = &Vuln{
				Title:   "html element injection",
				Param:   p,
				Req:     string(reqDump),
				Resp:    string(respDump),
				Attempt: payload.Attempt,
			}
			break
		}
	}
	return vuln, nil
}

func containsHTMLElement(rawHTML []byte, tagName string) bool {
	if !bytes.Contains(rawHTML, []byte(tagName)) {
		return false
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(rawHTML))
	if err != nil {
		return false
	}
	exists := false
	doc.Find(tagName).EachWithBreak(func(i int, s *goquery.Selection) bool { // goquery tagName selector is case-insensitive
		exists = true
		return false
	})
	return exists
}

func containsHTMLAttr(rawHTML []byte, attrName string) bool {
	if !bytes.Contains(rawHTML, []byte(attrName)) {
		return false
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(rawHTML))
	if err != nil {
		return false
	}
	exists := false
	doc.Find("[" + attrName + "]").EachWithBreak(func(i int, s *goquery.Selection) bool { // TODO: confirm it is case-sensitive or not
		exists = true
		return false
	})
	return exists
}

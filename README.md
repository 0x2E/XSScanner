# XSScanner

An XSS scanner demo based on parsed html nodes instead of regex.

## Usage

1. run vuln example:

```shell
go run example/*
```

2. run scanner:

```shell
go run . -u "http://localhost:9090/html-attr?id=1&file=2"
```

example:

```shell
$ go run . -u "http://localhost:9090/html-attr?id=1&file=2"
INFO[20230818 15:02:38] collecting params
INFO[20230818 15:02:38] found 2 params
INFO[20230818 15:02:38] found 2 vulns
INFO[20230818 15:02:38] {
	"url": "http://localhost:9090/html-attr?id=1\u0026file=2",
	"vuln": [
		{
			"title": "dangerous attr: src",
			"param": {
				"name": "file",
				"base_url": "http://localhost:9090/html-attr?id=1\u0026file=2",
				"type": "query",
				"form_type": ""
			},
			"req": "",
			"resp": "",
			"attempt": null
		},
		{
			"title": "html element injection",
			"param": {
				"name": "file",
				"base_url": "http://localhost:9090/html-attr?id=1\u0026file=2",
				"type": "query",
				"form_type": ""
			},
			"req": "GET /html-attr?file=%22%3EaHrefjavAScrIpTxMA%22%3Cx\u0026id=1 HTTP/1.1\r\nHost: localhost:9090\r\nUser-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36\r\nAccept-Encoding: gzip\r\n\r\n",
			"resp": "HTTP/1.1 200 OK\r\nContent-Length: 183\r\nContent-Type: text/html; charset=utf-8\r\nDate: Fri, 18 Aug 2023 07:02:38 GMT\r\n\r\n\n    \u003chtml\u003e\n        \u003cbody\u003e\n            \u003cdiv data-1=\"xxxxx\"\u003e\n                \u003cp\u003exxx\u003c/p\u003e\n            \u003c/div\u003e\n            \u003cimg src=\"\"\u003eaHrefjavAScrIpTxMA\"\u003cx\"\u003e\n        \u003c/body\u003e\n    \u003c/html\u003e\n\t",
			"attempt": [
				"\"\u003e\u003crqUVv/\u003e\"\u003cx",
				"\"\u003eaHrefjavAScrIpTxMA\"\u003cx"
			]
		}
	]
}
```

## ToDo

- headless Chrome
- parsing JS with tree-sitter
- testing with https://public-firing-range.appspot.com/

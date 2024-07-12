package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
)

type WebTest struct {
	testContext *testing.T
	req         *http.Request
	resp        *httptest.ResponseRecorder
	e           *echo.Echo
	ec          echo.Context
	doc         *goquery.Document
	failed      bool
}

func webTest(t *testing.T) *WebTest {
	wt := WebTest{
		testContext: t,
	}
	wt.req = httptest.NewRequest(http.MethodGet, "/", nil)
	wt.resp = httptest.NewRecorder()
	wt.e = echo.New()
	wt.e.Renderer = &Template{}
	wt.ec = wt.e.NewContext(wt.req, wt.resp)
	return &wt
}

func (wt *WebTest) confirmSuccessResponse() {
	if wt.resp.Code >= 400 {
		wt.failed = true
		wt.testContext.Errorf("got HTTP status code %d, expected 2xx or 3xx", wt.resp.Code)
	}
}

func (wt *WebTest) confirmBodyIncludes(query string, expected string) {
	if wt.doc == nil {
		wt.doc, _ = goquery.NewDocumentFromReader(bytes.NewReader(wt.resp.Body.Bytes()))
	}
	text := wt.doc.Find(query).Text()
	if !strings.Contains(text, expected) {
		wt.failed = true
		wt.testContext.Errorf("Did not find `%s` in %s", expected, query)
	}
}

func (wt *WebTest) showBodyOnFail() {
	if wt.failed {
		wt.testContext.Error(string(wt.resp.Body.Bytes()[:]))
	}
}

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

func (wt *WebTest) setParam(name string, value string) {
	wt.ec.SetParamNames(name)
	wt.ec.SetParamValues(value)
}

func (wt *WebTest) setQuery(name string, value string) {
	wt.ec.QueryParams().Add(name, value)
}

func (wt *WebTest) post(content string) {
	wt.req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(content))
	wt.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wt.ec = wt.e.NewContext(wt.req, wt.resp)
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

func (wt *WebTest) confirmRedirect(target string) {
	if wt.resp.Result().StatusCode != http.StatusSeeOther {
		wt.failed = true
		wt.testContext.Errorf("Did not get redirect status code %d", wt.resp.Result().StatusCode)
	}
	if wt.resp.Header().Get("Location") != target {
		wt.failed = true
		wt.testContext.Errorf("Unexpected redirect target: `%s`", wt.resp.Header().Get("Location"))
	}
}

func (wt *WebTest) showBodyOnFail() {
	if wt.failed {
		wt.testContext.Error(string(wt.resp.Body.Bytes()[:]))
	}
}

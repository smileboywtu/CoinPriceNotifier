// Copyright 2014 The GiterLab Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package urllib is a httplib for golang
package urllib

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var defaultSetting = HttpSettings{
	ShowDebug:        false,
	UserAgent:        "GiterLab",
	ConnectTimeout:   60 * time.Second,
	ReadWriteTimeout: 60 * time.Second,
	TlsClientConfig:  nil,
	Proxy:            nil,
	Transport:        nil,
	EnableCookie:     false,
	Gzip:             true,
	DumpBody:         true,
}
var defaultCookieJar http.CookieJar
var settingMutex sync.Mutex

// createDefaultCookie creates a global cookiejar to store cookies.
func createDefaultCookie() {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultCookieJar, _ = cookiejar.New(nil)
}

// Overwrite default settings
func SetDefaultSetting(setting HttpSettings) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultSetting = setting
	if defaultSetting.ConnectTimeout == 0 {
		defaultSetting.ConnectTimeout = 60 * time.Second
	}
	if defaultSetting.ReadWriteTimeout == 0 {
		defaultSetting.ReadWriteTimeout = 60 * time.Second
	}
}

// Get current default settings
func GetDefaultSetting() *HttpSettings {
	settingMutex.Lock()
	defer settingMutex.Unlock()

	return &defaultSetting
}

// return *HttpRequest with specific method
func newRequest(rawurl, method string) *HttpRequest {
	var resp http.Response
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Fatal(err)
	}

	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	return &HttpRequest{
		url:     rawurl,
		req:     &req,
		params:  map[string]string{},
		files:   map[string]string{},
		setting: defaultSetting,
		resp:    &resp,
		body:    nil,
	}
}

// Get returns *HttpRequest with GET method.
func Get(url string) *HttpRequest {
	return newRequest(url, "GET")
}

// Post returns *HttpRequest with POST method.
func Post(url string) *HttpRequest {
	return newRequest(url, "POST")
}

// Put returns *HttpRequest with PUT method.
func Put(url string) *HttpRequest {
	return newRequest(url, "PUT")
}

// Delete returns *HttpRequest DELETE method.
func Delete(url string) *HttpRequest {
	return newRequest(url, "DELETE")
}

// Head returns *HttpRequest with HEAD method.
func Head(url string) *HttpRequest {
	return newRequest(url, "HEAD")
}

// Head returns *HttpRequest with PATCH method.
func Patch(url string) *HttpRequest {
	return newRequest(url, "PATCH")
}

// HttpSettings
type HttpSettings struct {
	ShowDebug        bool
	UserAgent        string
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	TlsClientConfig  *tls.Config
	Proxy            func(*http.Request) (*url.URL, error)
	Transport        http.RoundTripper
	EnableCookie     bool
	Gzip             bool
	DumpBody         bool
}

// HttpRequest provides more useful methods for requesting one url than http.Request.
type HttpRequest struct {
	url     string
	req     *http.Request
	params  map[string]string
	files   map[string]string
	setting HttpSettings
	resp    *http.Response
	body    []byte
	dump    []byte
}

// Change request settings
func (b *HttpRequest) Setting(setting HttpSettings) *HttpRequest {
	b.setting = setting
	return b
}

// SetBasicAuth sets the request's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (b *HttpRequest) SetBasicAuth(username, password string) *HttpRequest {
	b.req.SetBasicAuth(username, password)
	return b
}

// SetEnableCookie sets enable/disable cookiejar
func (b *HttpRequest) SetEnableCookie(enable bool) *HttpRequest {
	b.setting.EnableCookie = enable
	return b
}

// SetUserAgent sets User-Agent header field
func (b *HttpRequest) SetUserAgent(useragent string) *HttpRequest {
	b.setting.UserAgent = useragent
	return b
}

// Debug sets show debug or not when executing request.
func (b *HttpRequest) Debug(isdebug bool) *HttpRequest {
	b.setting.ShowDebug = isdebug
	return b
}

// Dump Body.
func (b *HttpRequest) DumpBody(isdump bool) *HttpRequest {
	b.setting.DumpBody = isdump
	return b
}

// return the DumpRequest
func (b *HttpRequest) DumpRequest() []byte {
	return b.dump
}

// return the DumpRequest string
func (b *HttpRequest) DumpRequestString() string {
	return string(b.DumpRequest())
}

// SetTimeout sets connect time out and read-write time out for Request.
func (b *HttpRequest) SetTimeout(connectTimeout, readWriteTimeout time.Duration) *HttpRequest {
	b.setting.ConnectTimeout = connectTimeout
	b.setting.ReadWriteTimeout = readWriteTimeout
	return b
}

// SetTLSClientConfig sets tls connection configurations if visiting https url.
func (b *HttpRequest) SetTLSClientConfig(config *tls.Config) *HttpRequest {
	b.setting.TlsClientConfig = config
	return b
}

// Header add header item string in request.
func (b *HttpRequest) Header(key, value string) *HttpRequest {
	b.req.Header.Set(key, value)
	return b
}

// Set HOST
func (b *HttpRequest) SetHost(host string) *HttpRequest {
	b.req.Host = host
	return b
}

// Set the protocol version for incoming requests.
// Client requests always use HTTP/1.1.
func (b *HttpRequest) SetProtocolVersion(vers string) *HttpRequest {
	if len(vers) == 0 {
		vers = "HTTP/1.1"
	}

	major, minor, ok := http.ParseHTTPVersion(vers)
	if ok {
		b.req.Proto = vers
		b.req.ProtoMajor = major
		b.req.ProtoMinor = minor
	}

	return b
}

// SetCookie add cookie into request.
func (b *HttpRequest) SetCookie(cookie *http.Cookie) *HttpRequest {
	b.req.Header.Add("Cookie", cookie.String())
	return b
}

// Get default CookieJar
func GetDefaultCookieJar() http.CookieJar {
	return defaultCookieJar
}

// Set transport to
func (b *HttpRequest) SetTransport(transport http.RoundTripper) *HttpRequest {
	b.setting.Transport = transport
	return b
}

// Set http proxy
// example:
//
//	func(req *http.Request) (*url.URL, error) {
// 		u, _ := url.ParseRequestURI("http://127.0.0.1:8118")
// 		return u, nil
// 	}
func (b *HttpRequest) SetProxy(proxy func(*http.Request) (*url.URL, error)) *HttpRequest {
	b.setting.Proxy = proxy
	return b
}

// Param adds query param in to request.
// params build query string as ?key1=value1&key2=value2...
func (b *HttpRequest) Param(key, value string) *HttpRequest {
	b.params[key] = value
	return b
}

func (b *HttpRequest) PostFile(formname, filename string) *HttpRequest {
	b.files[formname] = filename
	return b
}

// Body adds request raw body.
// it supports string and []byte.
func (b *HttpRequest) Body(data interface{}) *HttpRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		b.req.Body = ioutil.NopCloser(bf)
		b.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		b.req.Body = ioutil.NopCloser(bf)
		b.req.ContentLength = int64(len(t))
	}
	return b
}

// JsonBody adds request raw body encoding by JSON.
func (b *HttpRequest) JsonBody(obj interface{}) (*HttpRequest, error) {
	if b.req.Body == nil && obj != nil {
		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		if err := enc.Encode(obj); err != nil {
			return b, err
		}
		b.req.Body = ioutil.NopCloser(buf)
		b.req.ContentLength = int64(buf.Len())
		b.req.Header.Set("Content-Type", "application/json")
	}
	return b, nil
}

func (b *HttpRequest) buildUrl(paramBody string) {
	// build GET url with query string
	if b.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Index(b.url, "?") != -1 {
			b.url += "&" + paramBody
		} else {
			b.url = b.url + "?" + paramBody
		}
		return
	}

	// build POST/PUT/PATCH url and body
	if (b.req.Method == "POST" || b.req.Method == "PUT" || b.req.Method == "PATCH") && b.req.Body == nil {
		// with files
		if len(b.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formname, filename := range b.files {
					fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
					if err != nil {
						log.Fatal(err)
					}
					fh, err := os.Open(filename)
					if err != nil {
						log.Fatal(err)
					}
					//iocopy
					_, err = io.Copy(fileWriter, fh)
					fh.Close()
					if err != nil {
						log.Fatal(err)
					}
				}
				for k, v := range b.params {
					bodyWriter.WriteField(k, v)
				}
				bodyWriter.Close()
				pw.Close()
			}()
			b.Header("Content-Type", bodyWriter.FormDataContentType())
			b.req.Body = ioutil.NopCloser(pr)
			return
		}

		// with params
		if len(paramBody) > 0 {
			b.Header("Content-Type", "application/x-www-form-urlencoded")
			b.Body(paramBody)
		}
	}
}

func (b *HttpRequest) getResponse() (*http.Response, error) {
	if b.resp.StatusCode != 0 {
		return b.resp, nil
	}
	resp, err := b.SendOut()
	if err != nil {
		return nil, err
	}
	b.resp = resp
	return resp, nil
}

func (b *HttpRequest) SendOut() (*http.Response, error) {
	var paramBody string
	if len(b.params) > 0 {
		var buf bytes.Buffer
		for k, v := range b.params {
			buf.WriteString(url.QueryEscape(k))
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
			buf.WriteByte('&')
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	b.buildUrl(paramBody)
	url, err := url.Parse(b.url)
	if err != nil {
		return nil, err
	}

	b.req.URL = url

	trans := b.setting.Transport

	if trans == nil {
		// create default transport
		trans = &http.Transport{
			TLSClientConfig: b.setting.TlsClientConfig,
			Proxy:           b.setting.Proxy,
			Dial:            TimeoutDialer(b.setting.ConnectTimeout, b.setting.ReadWriteTimeout),
		}
	} else {
		// if b.transport is *http.Transport then set the settings.
		if t, ok := trans.(*http.Transport); ok {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = b.setting.TlsClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = b.setting.Proxy
			}
			if t.Dial == nil {
				t.Dial = TimeoutDialer(b.setting.ConnectTimeout, b.setting.ReadWriteTimeout)
			}
		}
	}

	var jar http.CookieJar = nil
	if b.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}

	client := &http.Client{
		Transport: trans,
		Jar:       jar,
	}

	if b.setting.UserAgent != "" && b.req.Header.Get("User-Agent") == "" {
		b.req.Header.Set("User-Agent", b.setting.UserAgent)
	}

	if b.setting.ShowDebug {
		dump, err := httputil.DumpRequest(b.req, b.setting.DumpBody)
		if err != nil {
			log.Println(err.Error())
		}
		b.dump = dump
	}
	return client.Do(b.req)
}

// String returns the body string in response.
// it calls Response inner.
func (b *HttpRequest) String() (string, error) {
	data, err := b.Bytes()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Bytes returns the body []byte in response.
// it calls Response inner.
func (b *HttpRequest) Bytes() ([]byte, error) {
	if b.body != nil {
		return b.body, nil
	}
	resp, err := b.getResponse()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if b.setting.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		b.body, err = ioutil.ReadAll(reader)
	} else {
		b.body, err = ioutil.ReadAll(resp.Body)
	}
	return b.body, err
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (b *HttpRequest) ToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := b.getResponse()
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// ToJson returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (b *HttpRequest) ToJson(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXml returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (b *HttpRequest) ToXml(v interface{}) error {
	data, err := b.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// Response executes request client gets response mannually.
func (b *HttpRequest) Response() (*http.Response, error) {
	return b.getResponse()
}

// TimeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}

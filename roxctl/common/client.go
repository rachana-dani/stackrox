package common

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/stackrox/rox/pkg/errox"
	"github.com/stackrox/rox/pkg/utils"
	"github.com/stackrox/rox/roxctl/common/flags"
	"github.com/stackrox/rox/roxctl/common/logger"
	"golang.org/x/net/http2"
)

var (
	http1NextProtos = []string{"http/1.1", "http/1.0"}
)

// RoxctlHTTPClient abstracts all HTTP-related functionalities required within roxctl
type RoxctlHTTPClient interface {
	DoReqAndVerifyStatusCode(path string, method string, code int, body io.Reader) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
	NewReq(method string, path string, body io.Reader) (*http.Request, error)
}

type roxctlClientImpl struct {
	http        *http.Client
	a           Auth
	forceHTTP1  bool
	useInsecure bool
}

func getURL(path string) (string, error) {
	endpoint, usePlaintext, err := flags.EndpointAndPlaintextSetting()
	if err != nil {
		return "", errors.Wrap(err, "could not get endpoint")
	}
	scheme := "https"
	if usePlaintext {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/%s", scheme, endpoint, strings.TrimLeft(path, "/")), nil
}

// GetRoxctlHTTPClient returns a new instance of RoxctlHTTPClient with the given configuration
func GetRoxctlHTTPClient(timeout time.Duration, forceHTTP1 bool, useInsecure bool, log logger.Logger) (RoxctlHTTPClient, error) {
	tlsConf, err := tlsConfigForCentral(log)
	if err != nil {
		return nil, errors.Wrap(err, "instantiating TLS configuration for central")
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConf,
	}
	if forceHTTP1 {
		transport.TLSClientConfig.NextProtos = http1NextProtos
	} else {
		// There's no reason to not use HTTP/2, but we don't go out of our way to do so.
		if err := http2.ConfigureTransport(transport); err != nil {
			transport.TLSClientConfig.NextProtos = http1NextProtos
		}
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	auth, err := newAuth(log)
	if err != nil {
		return nil, err
	}
	return &roxctlClientImpl{http: client, a: auth, forceHTTP1: forceHTTP1, useInsecure: useInsecure}, nil
}

// DoReqAndVerifyStatusCode executes a http.Request and verifies that the http.Response had the given status code
func (client *roxctlClientImpl) DoReqAndVerifyStatusCode(path string, method string, code int, body io.Reader) (*http.Response, error) {
	req, err := client.NewReq(method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != code {
		defer utils.IgnoreError(resp.Body.Close)
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "Expected status code %d, but received %d. Additionally, there was an error reading the response", code, resp.StatusCode)
		}
		return nil, errox.InvariantViolation.Newf("expected status code %d, but received %d. Response Body: %s", code, resp.StatusCode, string(data))
	}

	return resp, nil
}

// DoHTTPRequestAndCheck200 does an http request to the provided path in Central,
// and passes through the remaining params. It checks that the returned status code is 200, and returns an error if it is not.
// The caller receives the http response object, which it is the caller's responsibility to close.
func DoHTTPRequestAndCheck200(path string, timeout time.Duration, method string, body io.Reader, log logger.Logger) (*http.Response, error) {
	client, err := GetRoxctlHTTPClient(timeout, flags.ForceHTTP1(), flags.UseInsecure(), log)
	if err != nil {
		return nil, err
	}
	return client.DoReqAndVerifyStatusCode(path, method, 200, body) //nolint:wrapcheck
}

// Do executes a http.Request
func (client *roxctlClientImpl) Do(req *http.Request) (*http.Response, error) {
	resp, err := client.http.Do(req)
	return resp, errors.Wrap(err, "error when doing http request")
}

// NewReq creates a new http.Request which will have all authentication metadata injected
func (client *roxctlClientImpl) NewReq(method string, path string, body io.Reader) (*http.Request, error) {
	reqURL, err := getURL(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, errors.Wrap(err, "error when creating http request")
	}
	if client.forceHTTP1 {
		req.ProtoMajor, req.ProtoMinor, req.Proto = 1, 1, "HTTP/1.1"
	}

	if req.URL.Scheme != "https" && !client.useInsecure {
		return nil, errox.InvalidArgs.Newf("URL %v uses insecure scheme %q, use --insecure flags to enable sending credentials", req.URL, req.URL.Scheme)
	}
	err = client.a.SetAuth(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not inject authentication information")
	}

	req.Header.Set("User-Agent", GetUserAgent())

	return req, nil
}

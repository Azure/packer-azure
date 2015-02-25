// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	neturl "net/url"
	"os"
	"strings"
	"time"

	"github.com/MSOpenTech/packer-azure/mod/pkg/crypto/tls"
	"github.com/MSOpenTech/packer-azure/mod/pkg/net/http"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/settings"
)

const apiVersion = "2014-06-01" // default API version

var logRequests = os.Getenv("PACKER_AZURE_NOLOG_REQUESTS") == ""

type Data struct {
	Verb string
	Uri  string
	Body []byte
}

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type Manager struct {
	SubscrId   string
	httpClient httpClient
}

func NewManager(publishSettingsPath, subscriptionName string) (*Manager, error) {
	subscriptionInfo, err := ParsePublishSettings(publishSettingsPath, subscriptionName)
	if err != nil {
		return nil, fmt.Errorf("Failed to Parse PublishSettings: %v", err)
	}

	var cert tls.Certificate

	cert, err = tls.X509KeyPair(subscriptionInfo.CertData, subscriptionInfo.CertData)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
		},
	}

	return &Manager{
		SubscrId:   subscriptionInfo.Id,
		httpClient: client,
	}, nil
}

func (m *Manager) Execute(req *Data) (resp *http.Response, err error) {
	resp, err = m.exec(req.Verb, req.Uri, req.Body, defaultRetryPolicy())
	return
}

func (m *Manager) ExecuteSync(req *Data) error {
	retryPolicy := defaultRetryPolicy()
	for {
		resp, err := m.exec(req.Verb, req.Uri, req.Body, retryPolicy)

		if err != nil {
			return err
		}

		errorMsg := fmt.Sprintf("Manager.ExecuteSync (%s %s): %%s", req.Verb, req.Uri)

		reqId, ok := resp.Header["X-Ms-Request-Id"]
		if !ok {
			return fmt.Errorf(errorMsg, "header key 'X-Ms-Request-Id' wasn't found")
		}

		log.Printf("Manager.ExecuteSync (%s %s) start polling for request ID %v", req.Verb, req.Uri, reqId[0])

		if _, err := m.WaitForOperation(reqId[0]); err != nil {
			if err, isAzureError := err.(model.AzureError); isAzureError {
				if shouldRetry, backoff := retryPolicy.ShouldRetry(err); shouldRetry {
					time.Sleep(backoff)
					continue
				}
				log.Printf("Manager.ExecuteSync (%s %s) caught non-retryable error: %v", req.Verb, req.Uri, err)
			}
			return err
		}
		break
	}
	return nil
}

// Exec executes REST request
func (d *Manager) exec(verb string, url string, body []byte, retryPolicy retryPolicy) (resp *http.Response, err error) {
	originalUrl := url // for HTTP 307 redirects

	if body == nil {
		body = make([]byte, 0)
	}

	for { // retry loop for azure errors
		url = originalUrl // reset url

		for { // retry loop for retry-able network errors
			if logRequests {
				log.Printf("Exec REQUEST: %s %s\n%s", verb, url, string(body))
			}
			req, err := http.NewRequest(verb, url, bytes.NewReader(body))
			if err != nil {
				return nil, err
			}
			req.Header.Add("Content-Type", "application/xml")
			req.Header.Add("x-ms-version", apiVersion)

			resp, err = d.httpClient.Do(req)

			if err != nil {
				switch err := err.(type) {
				case *neturl.Error:
					switch err := err.Err.(type) {
					case *net.OpError:
						switch {
						case err.Temporary(),
							err.Timeout(),
							strings.Contains(err.Err.Error(), "connection reset by peer"),
							strings.Contains(err.Err.Error(), "An existing connection was forcibly closed by the remote host"):
							log.Printf("Encountered retryable error: %+v", err)
							time.Sleep(500 * time.Millisecond)
							continue
						default:
							log.Printf("unhandled net.OpError (Op=%s, Net=%s, Addr=%+v, Err=%T, %+v)", err.Op, err.Net, err.Addr, err.Err, err.Err)
						}
					default:
						log.Printf("unhandled neturl.Error (Err=%+v, resp=%+v, %T)", err, resp, err)
					}
				default:
					log.Printf("unhandled error from httpClient.Do (Err=%+v, resp=%+v, %T)", err, resp, err)
				}

				return nil, err
			}
			break
		}

		if logRequests {
			buf := bytes.Buffer{}
			if _, err := buf.ReadFrom(resp.Body); err != nil {
				panic(err)
			}
			resp.Body.Close()
			resp.Body = ioutil.NopCloser(&buf)
			log.Printf("Exec RESPONSE: %s\nHeaders: %+v\n%s", resp.Status, resp.Header, string(buf.Bytes()))
		}

		if resp.StatusCode == 307 { // Temporary Redirect
			locationHeader, ok := resp.Header["Location"]
			if !ok {
				log.Printf("%s %s", "Failed to redirect:", "header key 'Location' wasn't found, retrying")
				continue
			}
			url = locationHeader[0]
			log.Printf("Redirecting: '%s' --> '%s'", originalUrl, url)
			continue
		}

		if resp.StatusCode >= 400 && resp.StatusCode <= 505 {

			defer resp.Body.Close()

			var respBody []byte
			respBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			if settings.LogRawResponseError {
				log.Printf("Exec raw error: %s\nHeaders: %+v\n%s", resp.Status, resp.Header, string(respBody))
			}

			azureError := model.AzureError{}
			err = xml.Unmarshal(respBody, &azureError)
			if err != nil {
				log.Printf("Exec could not unmarshal error: %v\n%s\nHeaders: %+v\n%s", err, resp.Status, resp.Header, string(respBody))
				return nil, err
			}

			if shouldRetry, backoff := retryPolicy.ShouldRetry(azureError); shouldRetry {
				time.Sleep(backoff)
				continue
			}

			return nil, azureError
		}

		break
	}

	return resp, err
}

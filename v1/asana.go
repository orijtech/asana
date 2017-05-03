// Copyright 2017 orijtech. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package asana

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

const baseURL = "https://app.asana.com/api/1.0"
const envAsanaPATKey = "ASANA_PERSONAL_ACCESS_TOKEN"

var (
	errEmptyEnvPATKey = fmt.Errorf("%q was not set in your environment", envAsanaPATKey)
)

// NewClient tries to use the first non-empty token passed otherwise
// if no tokens are passed in, it will look for the variable
//   `ASANA_PERSONAL_ACCESS_TOKEN`
// in your environment.
// It returns an error if it fails to find any API key to use.
func NewClient(personalAccessTokens ...string) (*Client, error) {
	pat := firstNonEmptyString(personalAccessTokens...)
	if pat == "" {
		pat = os.Getenv(envAsanaPATKey)
		if pat == "" {
			return nil, errEmptyEnvPATKey
		}
	}
	client := &Client{paToken: pat}
	return client, nil
}

func firstNonEmptyString(keys ...string) string {
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			return key
		}
	}
	return ""
}

type Client struct {
	paToken string
	sync.RWMutex

	rt http.RoundTripper
}

func (c *Client) SetHTTPRoundTripper(rt http.RoundTripper) {
	c.Lock()
	defer c.Unlock()
	c.rt = rt
}

func (c *Client) httpClient() *http.Client {
	c.RLock()
	defer c.RUnlock()

	rt := c.rt
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &http.Client{Transport: rt}
}

type User struct {
	UID UserID `json:"user"`
}

type UserID string

var _ json.Marshaler = (*UserID)(nil)

const MeAsUser = "me"

func (uid UserID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uid.String())
}

func (uid UserID) String() string {
	str := string(uid)
	if strings.TrimSpace(str) == "" {
		return MeAsUser
	}
	return str
}

var errUnimplemented = errors.New("unimplemented")

func (c *Client) personalAccessTokenAuthValue() string {
	c.RLock()
	defer c.RUnlock()

	return fmt.Sprintf("Bearer %s", c.paToken)
}

type HTTPError struct {
	msg  string
	code int
}

func (he HTTPError) Error() string {
	return he.msg
}

func (he HTTPError) Code() int {
	return he.code
}

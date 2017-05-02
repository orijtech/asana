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

func (c *Client) ListTasks() {
}

func (c *Client) DeleteTask() {
}

func (c *Client) UpdateTask() {
}

func (c *Client) Watch() {
}

const ()

type Team struct {
}

func (c *Client) CreateTeam() {
}

func (c *Client) FindTeamById() {
}

func (c *Client) FindMatchingTeams() {
}

type User struct {
	UID UserID `json:"user"`
}

type UserID string

var _ json.Marshaler = (*UserID)(nil)

const meAsUser = "me"

func (uid UserID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uid.String())
}

func (uid UserID) String() string {
	str := string(uid)
	if strings.TrimSpace(str) == "" {
		return meAsUser
	}
	return str
}

type TeamRequest struct {
	// TeamID is a globally unique identifier for the team.
	TeamID string `json:"team_id"`

	UserID string `json:"user_id"`

	OrganizationID string `json:"organization"`
}

var (
	errNilTeamRequest = errors.New("nil team request")

	errEmptyUserID = errors.New("empty userID passed in")
	errEmptyTeamID = errors.New("empty teamID passed in")
)

func (treq *TeamRequest) Validate() error {
	if treq == nil {
		return errNilTeamRequest
	}
	teamID := strings.TrimSpace(treq.TeamID)
	if teamID == "" {
		return errEmptyTeamID
	}
	if treq.UserID == "" {
		return errEmptyUserID
	}
	return nil
}

type teamCRUDData struct {
	UserID UserID `json:"user"`
}

var errUnimplemented = errors.New("unimplemented")

// POST: "/teams/{TEAMID}/addUser"
func (c *Client) AddUserToTeam(treq *TeamRequest) (interface{}, error) {
	if err := treq.Validate(); err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/teams/%s/addUser", baseURL, treq.TeamID)
	if fullURL == "" {
	}
	return nil, errUnimplemented
}

func (c *Client) RemoveUserFromTeam(treq *TeamRequest) (interface{}, error) {
	if err := treq.Validate(); err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/teams/%s/deleteUser", baseURL, treq.TeamID)
	if fullURL == "" {
	}
	return nil, errUnimplemented
}

func (c *Client) UsersInTeam(treq *TeamRequest) (interface{}, error) {
	if treq == nil || strings.TrimSpace(treq.TeamID) == "" {
		return nil, errEmptyTeamID
	}
	fullURL := fmt.Sprintf("%s/teams/%s/users", baseURL, treq.TeamID)
	if fullURL == "" {
	}
	return nil, errUnimplemented
}

func (c *Client) TeamsForUser(treq *TeamRequest) (interface{}, error) {
	if treq == nil || strings.TrimSpace(treq.UserID) == "" {
		return nil, errEmptyUserID
	}
	fullURL := fmt.Sprintf("%s/users/%s/team", baseURL, treq.UserID)
	req, _ := http.NewRequest("GET", fullURL, nil)
	if req == nil {
	}
	return nil, errUnimplemented
}

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

package asana

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/orijtech/otils"
)

type Layout string

const (
	ListLayout  Layout = "list"
	BoardLayout Layout = "board"
)

var (
	_ json.Unmarshaler = (*Layout)(nil)
	_ json.Marshaler   = (*Layout)(nil)
)

func (l *Layout) MarshalJSON() ([]byte, error) {
	str := string(ListLayout)
	if l != nil {
		str = string(*l)
	}
	b := []byte(strconv.Quote(str))
	return b, nil
}

func (l *Layout) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	*l = Layout(unquoted)
	return nil
}

type ProjectRequest struct {
	Name  string `json:"name,omitempty"`
	Notes string `json:"notes,omitempty"`

	Color  string `json:"color,omitempty"`
	Layout Layout `json:"layout,omitempty"`

	Team *NamedAndIDdEntity `json:"team,omitempty"`

	Workspace string `json:"workspace,omitempty"`

	PublicToOrganization bool `json:"public,omitempty"`
}

type Project struct {
	Name     string `json:"name,omitempty"`
	Notes    string `json:"notes,omitempty"`
	Color    string `json:"color,omitempty"`
	Archived bool   `json:"archived,omitempty"`

	Owner      *NamedAndIDdEntity `json:"owner,omitempty"`
	CreatedAt  *time.Time         `json:"created_at,omitempty"`
	ModifiedAt *time.Time         `json:"created_at,omitempty"`

	WorkspaceID int64 `json:"workspace,string"`

	Members   []*NamedAndIDdEntity `json:"members,omitempty"`
	Followers []*NamedAndIDdEntity `json:"followers,omitempty"`
}

var (
	errNilProject     = errors.New("nil project")
	errEmptyWorkspace = errors.New("expecting a non-empty workspace")
)

func (preq *ProjectRequest) Validate() error {
	if preq == nil {
		return errNilProject
	}
	if preq.Workspace == "" {
		return errEmptyWorkspace
	}
	return nil
}

type projectWrap struct {
	Project *Project `json:"data"`
}

func parseOutProjectFromData(blob []byte) (*Project, error) {
	log.Printf("%s\n", blob)
	pwj := new(projectWrap)
	if err := json.Unmarshal(blob, pwj); err != nil {
		return nil, err
	}
	return pwj.Project, nil
}

func (c *Client) CreateProject(preq *ProjectRequest) (*Project, error) {
	if err := preq.Validate(); err != nil {
		return nil, err
	}

	qs, err := otils.ToURLValues(preq)
	if err != nil {
		return nil, err
	}

	queryStr := qs.Encode()
	fullURL := fmt.Sprintf("%s/projects", baseURL)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(queryStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}
	return parseOutProjectFromData(slurp)
}

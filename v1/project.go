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
	ProjectID  string `json:"id"`
	ProjectGID string `json:"project_gid"`

	Name  string `json:"name,omitempty"`
	Notes string `json:"notes,omitempty"`

	Color  string `json:"color,omitempty"`
	Layout Layout `json:"layout,omitempty"`

	Team *NamedAndIDdEntity `json:"team,omitempty"`

	Workspace string `json:"workspace,omitempty"`

	PublicToOrganization bool     `json:"public,omitempty"`
	Members              []string `json:"members,omitempty"`
}

type Project struct {
	ID       int64              `json:"id,omitempty"`
	GID      string             `json:"gid,omitempty"`
	Team     *NamedAndIDdEntity `json:"team,omitempty"`
	Name     string             `json:"name,omitempty"`
	Notes    string             `json:"notes,omitempty"`
	Color    string             `json:"color,omitempty"`
	Archived bool               `json:"archived,omitempty"`

	Owner      *NamedAndIDdEntity `json:"owner,omitempty"`
	CreatedAt  *time.Time         `json:"created_at,omitempty"`
	ModifiedAt *time.Time         `json:"created_at,omitempty"`

	Workspace *NamedAndIDdEntity `json:"workspace,omitempty"`

	Members   []*NamedAndIDdEntity `json:"members,omitempty"`
	Followers []*NamedAndIDdEntity `json:"followers,omitempty"`
}

var (
	errNilProjectRequest = errors.New("expecting a non-nil projectRequest")
	errEmptyWorkspace    = errors.New("expecting a non-empty workspace")
)

func (preq *ProjectRequest) Validate() error {
	if preq == nil {
		return errNilProjectRequest
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
	pwj := new(projectWrap)
	if err := json.Unmarshal(blob, pwj); err != nil {
		return nil, err
	}
	return pwj.Project, nil
}

var errImmutableWorkspace = errors.New("workspace once set cannot be modified")

// UpdateProject changes the attributes of a project.
// Note that some fields like Workspace cannot be changed
// once the project has been created. Trying to modify this
// field will return an error.
func (c *Client) UpdateProject(preq *ProjectRequest) (*Project, error) {
	if preq == nil {
		return nil, errNilProjectRequest
	}
	projectID := strings.TrimSpace(preq.ProjectID)
	if projectID == "" {
		return nil, errEmptyProjectID
	}
	if preq.Workspace != "" {
		return nil, errImmutableWorkspace
	}

	copyReq := *preq
	// Now unset ProjectID to avoid problems
	// with trying to mutate it on the backend.
	copyReq.ProjectID = ""

	qs, err := otils.ToURLValues(&copyReq)
	if err != nil {
		return nil, err
	}

	queryStr := qs.Encode()
	fullURL := fmt.Sprintf("%s/projects/%s", baseURL, projectID)
	req, err := http.NewRequest("PUT", fullURL, strings.NewReader(queryStr))
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

func (c *Client) AddUsersToProject(preq *ProjectRequest) error {
	if err := preq.Validate(); err != nil {
		return err
	}

	qs, err := otils.ToURLValues(preq)
	if err != nil {
		return err
	}

	queryStr := qs.Encode()
	fullURL := fmt.Sprintf("%s/projects/%s/addMembers", baseURL, preq.ProjectGID)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(queryStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, _, err = c.doAuthReqThenSlurpBody(req)
	return err
}

func (c *Client) RemoveUsersFromProject(preq *ProjectRequest) error {
	if err := preq.Validate(); err != nil {
		return err
	}

	qs, err := otils.ToURLValues(preq)
	if err != nil {
		return err
	}

	queryStr := qs.Encode()
	fullURL := fmt.Sprintf("%s/projects/%s/removeMembers", baseURL, preq.ProjectGID)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(queryStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, _, err = c.doAuthReqThenSlurpBody(req)
	return err
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

func (c *Client) FindProjectByID(projectID string) (*Project, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, errEmptyProjectID
	}
	fullURL := fmt.Sprintf("%s/projects/%s", baseURL, projectID)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}
	return parseOutProjectFromData(slurp)
}

func (c *Client) DeleteProjectByID(projectID string) error {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return errEmptyProjectID
	}
	fullURL := fmt.Sprintf("%s/projects/%s", baseURL, projectID)
	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return err
	}
	_, _, err = c.doAuthReqThenSlurpBody(req)
	return err
}

type ProjectQuery struct {
	WorkspaceID string `json:"workspace,omitempty"`
	TeamID      string `json:"team,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
}

var errNilProjectQuery = errors.New("expecting a non-nil projectQuery")

type ProjectsPage struct {
	Projects []*Project `json:"data"`
	Err      error
}

type projectsPager struct {
	ProjectsPage

	NextPage *pageToken `json:"next_page,omitempty"`
}

// FindProjects queries for projects with atleast one
// of the fields of the ProjectQuery set as a filter.
func (c *Client) QueryForProjects(pq *ProjectQuery) (pagesChan chan *ProjectsPage, cancelChan chan<- bool, err error) {
	if pq == nil {
		return nil, nil, errNilProjectQuery
	}
	qs, err := otils.ToURLValues(pq)
	if err != nil {
		return nil, nil, err
	}

	cancelChan = make(chan bool, 1)
	pagesChan = make(chan *ProjectsPage)

	go func() {
		defer close(pagesChan)

		path := fmt.Sprintf("/projects?opt_fields=id,gid,name,notes,team,members&%s", qs.Encode())
		for {
			fullURL := fmt.Sprintf("%s%s", baseURL, path)
			req, _ := http.NewRequest("GET", fullURL, nil)
			slurp, _, err := c.doAuthReqThenSlurpBody(req)
			if err != nil {
				pagesChan <- &ProjectsPage{Err: err}
				return
			}

			page := new(projectsPager)
			if err := json.Unmarshal(slurp, page); err != nil {
				page.Err = err
			}

			pp := page.ProjectsPage
			pagesChan <- &pp

			if np := page.NextPage; np != nil && np.Path == "" {
				path = np.Path
			} else {
				// End of this pagination
				break
			}
		}
	}()

	return pagesChan, cancelChan, nil
}

func (c *Client) TasksForProject(projectID string) (resultsChan chan *TaskResultPage, cancelChan chan<- bool, err error) {
	if projectID == "" {
		return nil, nil, errEmptyProjectID
	}

	startPath := fmt.Sprintf("/projects/%s/tasks", projectID)
	return c.doTasksPaging(startPath)
}

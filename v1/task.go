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
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/orijtech/otils"
)

type Task struct {
	ID          int64              `json:"id,omitempty"`
	Assignee    *NamedAndIDdEntity `json:"assignee,omitempty"`
	CreatedAt   *time.Time         `json:"created_at,omitempty"`
	Completed   bool               `json:"completed,omitempty"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`

	AssigneeStatus AssigneeStatus `json:"assignee_status,omitempty"`

	CustomFields []CustomField `json:"custom_fields,omitempty"`

	DueOn *YYYYMMDD  `json:"due_on,omitempty"`
	DueAt *time.Time `json:"due_at,omitempty"`

	Metadata Metadata `json:"external,omitempty"`

	Followers []*NamedAndIDdEntity `json:"followers,omitempty"`

	HeartedByMe bool                 `json:"hearted,omitempty"`
	Hearts      []*NamedAndIDdEntity `json:"hearts,omitempty"`
	HeartCount  int64                `json:"num_hearts,omitempty"`
	ModifiedAt  *time.Time           `json:"modified_at"`

	Name string `json:"name,omitempty"`

	Notes string `json:"notes,omitempty"`

	Projects   []*Project `json:"projects,omitempty"`
	ParentTask *Task      `json:"parent,omitempty"`

	Workspace *NamedAndIDdEntity `json:"workspace,omitempty"`

	Memberships []*Membership `json:"memberships,omitempty"`

	Tags []*NamedAndIDdEntity `json:"tags,omitempty"`
}

type NamedAndIDdEntity struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type Membership struct {
	Project *NamedAndIDdEntity `json:"project,omitempty"`
	Section *NamedAndIDdEntity `json:"section,omitempty"`
}

type AssigneeStatus string

const (
	StatusInbox    AssigneeStatus = "inbox"
	StatusLater    AssigneeStatus = "later"
	StatusToday    AssigneeStatus = "today"
	StatusUpcoming AssigneeStatus = "upcoming"
)

const (
	defaultAssigneeStatus = StatusInbox
)

func (as AssigneeStatus) String() string {
	str := string(as)
	if str != "" {
		return str
	}
	return string(defaultAssigneeStatus)
}

func (as AssigneeStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(as))
}

type CustomField map[string]interface{}

type Metadata map[string]interface{}

type YYYYMMDD struct {
	sync.RWMutex

	YYYY int64
	MM   int64
	DD   int64

	str string
}

var _ json.Marshaler = (*YYYYMMDD)(nil)
var _ json.Unmarshaler = (*YYYYMMDD)(nil)

func (ymd *YYYYMMDD) UnmarshalJSON(b []byte) error {
	// Format of data is: 2012-03-26
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	splits := strings.Split(unquoted, "-")
	if len(splits) < 3 {
		return errors.New("expecting YYYY-MM-DD")
	}

	var intified []int64
	for _, split := range splits {
		it, err := intifyIt(split)
		if err != nil {
			return err
		}
		intified = append(intified, it)
	}

	ymd.YYYY = intified[0]
	ymd.MM = intified[1]
	ymd.DD = intified[2]
	return nil
}

func (ymd *YYYYMMDD) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ymd.String()))
}

func (ymd *YYYYMMDD) String() string {
	if ymd == nil {
		return ""
	}
	ymd.Lock()
	defer ymd.Unlock()
	if ymd.str == "" {
		ymd.str = fmt.Sprintf("%d-%d-%d", ymd.YYYY, ymd.MM, ymd.DD)
	}
	return ymd.str
}

func intifyIt(st string) (int64, error) {
	return strconv.ParseInt(st, 10, 64)
}

func (c *Client) doAuthReqThenSlurpBody(req *http.Request) ([]byte, http.Header, error) {
	req.Header.Set("Authorization", c.personalAccessTokenAuthValue())
	res, err := c.httpClient().Do(req)
	if err != nil {
		return nil, nil, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	if !otils.StatusOK(res.StatusCode) {
		errMsg := res.Status
		if res.Body != nil {
			slurp, _ := ioutil.ReadAll(res.Body)
			if len(slurp) > 0 {
				errMsg = string(slurp)
			}
		}
		return nil, res.Header, &HTTPError{msg: errMsg, code: res.StatusCode}
	}

	slurp, err := ioutil.ReadAll(res.Body)
	return slurp, res.Header, err
}

var readOnlyFields = []string{
	"num_hearts",
}

type taskResultWrap struct {
	Task *Task `json:"data"`
}

func (c *Client) CreateTask(t *TaskRequest) (*Task, error) {
	// This endpoint takes in url-encoded data
	qs, err := otils.ToURLValues(t)
	if err != nil {
		return nil, err
	}

	for _, field := range readOnlyFields {
		qs.Del(field)
	}

	fullURL := fmt.Sprintf("%s/tasks", baseURL)
	queryStr := qs.Encode()
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(queryStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}
	return parseOutTaskFromData(slurp)
}

func parseOutTaskFromData(blob []byte) (*Task, error) {
	wrap := new(taskResultWrap)
	if err := json.Unmarshal(blob, wrap); err != nil {
		return nil, err
	}
	return wrap.Task, nil
}

type TaskResultPage struct {
	Tasks []*Task `json:"data"`
	Err   error
}

type taskPager struct {
	TaskResultPage

	NextPage *pageToken `json:"next_page,omitempty"`
}

type TaskRequest struct {
	Page        int        `json:"page,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	MaxRetries  int        `json:"max_retries,omitempty"`
	Assignee    string     `json:"assignee"`
	ProjectID   string     `json:"project,omitempty"`
	Workspace   string     `json:"workspace,omitempty"`
	ID          int64      `json:"id,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	Completed   bool       `json:"completed,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	AssigneeStatus AssigneeStatus `json:"assignee_status,omitempty"`

	CustomFields []CustomField `json:"custom_fields,omitempty"`

	DueOn *YYYYMMDD  `json:"due_on,omitempty"`
	DueAt *time.Time `json:"due_at,omitempty"`

	Metadata Metadata `json:"external,omitempty"`

	Followers []UserID `json:"followers,omitempty"`

	HeartedByMe bool       `json:"hearted,omitempty"`
	Hearts      []*User    `json:"hearts,omitempty"`
	HeartCount  int64      `json:"num_hearts,omitempty"`
	ModifiedAt  *time.Time `json:"modified_at"`

	Name string `json:"name,omitempty"`

	Notes string `json:"notes,omitempty"`

	Projects   []*NamedAndIDdEntity `json:"projects,omitempty"`
	ParentTask *Task                `json:"parent,omitempty"`

	Memberships []*Membership `json:"memberships,omitempty"`

	Tags []*NamedAndIDdEntity `json:"tags,omitempty"`
}

type listTaskWrap struct {
	Tasks []*Task `json:"data"`
}

func (c *Client) ListAllMyTasks() (resultsChan chan *TaskResultPage, cancelChan chan<- bool, err error) {
	cancelChan = make(chan bool)
	treq, err := c.ListMyTasks(nil)
	return treq, cancelChan, err
}

const defaultTaskLimit = 20

func (treq *TaskRequest) fillWithDefaults() {
	if treq == nil {
		return
	}
	if treq.Limit <= 0 {
		treq.Limit = defaultTaskLimit
	}
}

func (c *Client) ListMyTasks(treq *TaskRequest) (chan *TaskResultPage, error) {
	theReq := new(TaskRequest)
	if treq != nil {
		*theReq = *treq
	}
	theReq.Assignee = MeAsUser
	theReq.fillWithDefaults()
	qs, err := otils.ToURLValues(theReq)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/tasks?%s", qs.Encode())
	pageChan, _, err := c.doTasksPaging(path)
	return pageChan, err
}

type WorkspacePage struct {
	Err        error
	Workspaces []*Workspace `json:"data,omitempty"`

	NextPage *pageToken `json:"next_page,omitempty"`
}

type pageToken struct {
	Offset string `json:"offset"`
	Path   string `json:"path"`
	URI    string `json:"uri"`
}

type Workspace NamedAndIDdEntity

func (c *Client) ListMyWorkspaces() (chan *WorkspacePage, error) {
	wspChan := make(chan *WorkspacePage)
	go func() {
		defer close(wspChan)

		path := "/workspaces"
		for {
			fullURL := fmt.Sprintf("%s%s", baseURL, path)
			req, _ := http.NewRequest("GET", fullURL, nil)
			slurp, _, err := c.doAuthReqThenSlurpBody(req)
			if err != nil {
				wspChan <- &WorkspacePage{Err: err}
				return
			}

			page := new(WorkspacePage)
			if err := json.Unmarshal(slurp, page); err != nil {
				page.Err = err
			}

			wspChan <- page

			if np := page.NextPage; np != nil && np.Path == "" {
				path = np.Path
			} else {
				// End of this pagination
				break
			}
		}
	}()

	return wspChan, nil
}

var errEmptyTaskID = errors.New("expecting a non-empty taskID")

func (c *Client) FindTaskByID(taskID string) (*Task, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errEmptyTaskID
	}
	fullURL := fmt.Sprintf("%s/tasks/%s", baseURL, taskID)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}
	return parseOutTaskFromData(slurp)
}

var errEmptyProjectID = errors.New("expecting a non-empty projectID")

func (c *Client) ListTasksForProject(treq *TaskRequest) (resultsChan chan *TaskResultPage, cancelChan chan<- bool, err error) {
	path := fmt.Sprintf("/projects/%s/tasks", treq.ProjectID)
	return c.doTasksPaging(path)
}

func (c *Client) doTasksPaging(path string) (resultsChan chan *TaskResultPage, cancelChan chan<- bool, err error) {
	tasksPageChan := make(chan *TaskResultPage)
	cancelChan = make(chan bool, 1)

	go func() {
		defer close(tasksPageChan)

		for {
			fullURL := fmt.Sprintf("%s%s", baseURL, path)
			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				tasksPageChan <- &TaskResultPage{Err: err}
				return
			}

			slurp, _, err := c.doAuthReqThenSlurpBody(req)
			if err != nil {
				tasksPageChan <- &TaskResultPage{Err: err}
				return
			}

			pager := new(taskPager)
			if err := json.Unmarshal(slurp, pager); err != nil {
				pager.Err = err
			}

			taskPage := pager.TaskResultPage
			tasksPageChan <- &taskPage

			if np := pager.NextPage; np != nil && np.Path == "" {
				path = np.Path
			} else {
				// End of this pagination
				break
			}
		}
	}()

	return tasksPageChan, cancelChan, nil
}

func (c *Client) DeleteTask(taskID string) error {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return errEmptyTaskID
	}
	fullURL := fmt.Sprintf("%s/tasks/%s", baseURL, taskID)
	req, _ := http.NewRequest("DELETE", fullURL, nil)
	_, _, err := c.doAuthReqThenSlurpBody(req)
	return err
}

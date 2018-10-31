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
	"strings"

	"github.com/orijtech/otils"
)

type teamCRUDData struct {
	UserID UserID `json:"user"`
}

type Team struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
	HtmlDescription string `json:"html_description"`
}

type TeamRequest struct {
	// TeamID is a globally unique identifier for the team.
	TeamID string `json:"team_id"`

	UserID string `json:"user_id"`

	OrganizationID string `json:"organization"`
}

var (
	errNilTeamRequest = errors.New("expecting a non-nil team request")

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

func (c *Client) AddUserToTeam(treq *TeamRequest) (*Team, error) {
	if err := treq.Validate(); err != nil {
		return nil, err
	}

	qs, err := otils.ToURLValues(treq)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s/teams/%s/addUser?user=%s", baseURL, treq.TeamID, treq.UserID)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(qs.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}

	tw := new(teamWrap)
	if err := json.Unmarshal(slurp, tw); err != nil {
		return nil, err
	}
	return tw.Team, nil
}

func (c *Client) RemoveUserFromTeam(treq *TeamRequest) error {
	if err := treq.Validate(); err != nil {
		return err
	}

	qs, err := otils.ToURLValues(treq)
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s/teams/%s/removeUser?user=%s", baseURL, treq.TeamID, treq.UserID)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(qs.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, _, err = c.doAuthReqThenSlurpBody(req)
	return err
}

type teamWrap struct {
	Team *Team `json:"data"`
}

func (c *Client) FindTeamByID(teamID string) (*Team, error) {
	if teamID == "" {
		return nil, errEmptyTeamID
	}
	fullURL := fmt.Sprintf("%s/teams/%s", baseURL, teamID)
	req, _ := http.NewRequest("GET", fullURL, nil)
	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}
	twrap := new(teamWrap)
	if err := json.Unmarshal(slurp, twrap); err != nil {
		return nil, err
	}
	return twrap.Team, nil
}

var errEmptyOrganizationID = errors.New("expecting a non-empty organizationID")

type TeamPage struct {
	Teams []*Team `json:"data"`
	Err   error
}

type teamPager struct {
	TeamPage

	NextPage *pageToken `json:"next_page,omitempty"`
}

func (c *Client) ListAllTeamsInOrganization(organizationID string) (pagesChan chan *TeamPage, cancelChan chan<- bool, err error) {
	if organizationID == "" {
		return nil, nil, errEmptyOrganizationID
	}

	startingPath := fmt.Sprintf("/organizations/%s/teams?opt_fields=html_description,name,id", organizationID)
	return c.pageForTeams(startingPath)
}

func (c *Client) ListAllTeamsForUser(treq *TeamRequest) (pagesChan chan *TeamPage, cancelChan chan<- bool, err error) {
	if treq == nil {
		return nil, nil, errNilTeamRequest
	}

	theUserID := treq.UserID
	if theUserID == "" {
		return nil, nil, errEmptyUserID
	}

	qs, err := otils.ToURLValues(treq)
	if err != nil {
		return nil, nil, err
	}

	startingPath := fmt.Sprintf("/users/%s/teams?%s", theUserID, qs.Encode())
	return c.pageForTeams(startingPath)
}

func (c *Client) pageForTeams(path string) (pagesChan chan *TeamPage, cancelChan chan<- bool, err error) {
	pagesChan = make(chan *TeamPage)
	cancelChan = make(chan bool, 1)

	go func() {
		defer close(pagesChan)

		for {
			fullURL := fmt.Sprintf("%s%s", baseURL, path)
			req, _ := http.NewRequest("GET", fullURL, nil)
			slurp, _, err := c.doAuthReqThenSlurpBody(req)
			if err != nil {
				pagesChan <- &TeamPage{Err: err}
				return
			}

			pager := new(teamPager)
			if err := json.Unmarshal(slurp, pager); err != nil {
				pager.Err = err
			}

			teamPage := pager.TeamPage
			pagesChan <- &teamPage

			if np := pager.NextPage; np != nil && np.Path == "" {
				path = np.Path
			} else {
				// End of this pagination
				break
			}
		}
	}()

	return pagesChan, cancelChan, nil
}

type UsersPage struct {
	Users []*User `json:"data"`
	Err   error
}

type usersPager struct {
	UsersPage

	NextPage *pageToken `json:"next_page,omitempty"`
}

func (c *Client) ListAllUsersInTeam(teamID string) (pagesChan chan *UsersPage, cancelChan chan<- bool, err error) {
	if teamID == "" {
		return nil, nil, errEmptyTeamID
	}

	cancelChan = make(chan bool, 1)
	pagesChan = make(chan *UsersPage)

	go func() {
		defer close(pagesChan)

		path := fmt.Sprintf("/teams/%s/users", teamID)
		for {
			fullURL := fmt.Sprintf("%s%s", baseURL, path)
			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				pagesChan <- &UsersPage{Err: err}
				return
			}
			slurp, _, err := c.doAuthReqThenSlurpBody(req)
			if err != nil {
				pagesChan <- &UsersPage{Err: err}
				return
			}

			pager := new(usersPager)
			if err := json.Unmarshal(slurp, pager); err != nil {
				pager.Err = err
			}
			usersPage := pager.UsersPage
			pagesChan <- &usersPage

			if np := pager.NextPage; np != nil && np.Path == "" {
				path = np.Path
			} else {
				// End of this pagination
				break
			}
		}
	}()

	return pagesChan, cancelChan, nil
}

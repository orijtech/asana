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

package asana_test

import (
	"fmt"
	"log"
	"os"

	"github.com/orijtech/asana/v1"
)

func Example_client_CreateTask() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	setupServers, err := client.CreateTask(&asana.TaskRequest{
		Assignee:  "emm.odeke@gmail.com",
		Notes:     "Announce Asana Go API client release",
		Name:      "api-client-release",
		Workspace: "331783765164429",
		Followers: []asana.UserID{
			"emmanuel@orijtech.com",
		},
	})

	if err != nil {
		log.Fatalf("the error: %#v", err)
	}

	log.Printf("Here is the task: %#v", setupServers)
}

func Example_client_ListMyTasks() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	taskPagesChan, err := client.ListMyTasks(&asana.TaskRequest{
		Workspace: "331727068525363",
	})
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range taskPagesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, task := range page.Tasks {
			log.Printf("Page: #%d i: %d task: %#v", pageCount, i, task)
		}
		pageCount += 1
	}
}

func Example_client_ListMyWorkspaces() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	workspacesChan, err := client.ListMyWorkspaces()
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range workspacesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, workspace := range page.Workspaces {
			log.Printf("Page: #%d i: %d workspace: %#v", pageCount, i, workspace)
		}
		pageCount += 1
	}
}

func Example_client_FindTaskByID() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	setupServers, err := client.FindTaskByID("332508471165497")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("TaskName: %s\n", setupServers.Name)
	if setupServers.HeartedByMe {
		fmt.Printf("I heart'd this task!\n")
	}

	fmt.Printf("Assignee: %#v\n", setupServers.AssigneeStatus)
	fmt.Printf("Assignee: %#v\n", setupServers.Assignee)
	fmt.Printf("Notes: %#v\n", setupServers.Notes)
	fmt.Printf("Followers\n")
	fmt.Printf("CreatedAt: %v\n", setupServers.CreatedAt)
	fmt.Printf("ModifiedAt: %v\n", setupServers.ModifiedAt)

	for _, follower := range setupServers.Followers {
		fmt.Printf("ID: %v Name: %s\n", follower.ID, follower.Name)
	}

	for i, heart := range setupServers.Hearts {
		fmt.Printf("#%d HeartID: %v Name: %s\n", i+1, heart.ID, heart.Name)
	}

	for _, tag := range setupServers.Tags {
		fmt.Printf("Tag: %v ID: %v\n", tag.Name, tag.ID)
	}
}

func Example_client_DeleteTask() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := client.DeleteTask("332508471165497"); err != nil {
		log.Fatalf("Task deletion err: %v", err)
	} else {
		log.Printf("Successfully deleted the task!")
	}
}

func Example_client_ListTasksForProject() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	taskPagesChan, _, err := client.ListTasksForProject(&asana.TaskRequest{
		ProjectID: "331783765164429",
	})
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range taskPagesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, task := range page.Tasks {
			log.Printf("Page: #%d i: %d task: %#v", pageCount, i, task)
		}
		pageCount += 1
	}
}

func Example_client_CreateProject() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	proj, err := client.CreateProject(&asana.ProjectRequest{
		Name:      "Project-Go",
		Notes:     "This is a port of api clients to Go",
		Layout:    asana.BoardLayout,
		Workspace: "331783765164429",

		PublicToOrganization: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created project: %#v", proj)
}

func Example_client_FindProjectByID() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	proj, err := client.FindProjectByID("332697649493087")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("The project: %#v", proj)
}

func Example_client_UpdateProject() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	proj, err := client.UpdateProject(&asana.ProjectRequest{
		ProjectID: "332697649493087",
		Name:      "Project-Go updated",
		Notes:     "We need to prioritize which features will be included\nAm also changing it to a list layout",
		Layout:    asana.ListLayout,

		PublicToOrganization: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Updated project: %#v", proj)
}

func Example_client_DeleteProject() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	projectID := "332697649493087"
	if err := client.DeleteProjectByID(projectID); err != nil {
		log.Printf("Successfully deleted project %q!", projectID)
	} else {
		log.Fatalf("Failed to delete project %q!", projectID)
	}
}

func Example_client_QueryForProjects() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	pagesChan, _, err := client.QueryForProjects(&asana.ProjectQuery{
		Archived:    false,
		WorkspaceID: "331783765164429",
	})

	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range pagesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, project := range page.Projects {
			log.Printf("Page: #%d i: %d project: %#v", pageCount, i, project)
		}
		pageCount += 1
	}
}

func Example_client_TasksForProject() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	tasksPagesChan, _, err := client.TasksForProject("332697157202049")
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range tasksPagesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, task := range page.Tasks {
			log.Printf("Page: #%d i: %d task: %#v", pageCount, i, task)
		}
		pageCount += 1
	}
}

func Example_client_FindTeamByID() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	engTeam, err := client.FindTeamByID("")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("This is the information for the 331783765164429 team: %#v", engTeam)
}

func Example_client_ListAllTeamsInOrganization() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	teamsPagesChan, _, err := client.ListAllTeamsInOrganization("332697157202049")
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range teamsPagesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, team := range page.Teams {
			log.Printf("Page: #%d i: %d team: %#v", pageCount, i, team)
		}
		pageCount += 1
	}
}

func Example_client_ListAllTeamsForUser() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	teamsPagesChan, _, err := client.ListAllTeamsForUser(&asana.TeamRequest{
		UserID:         asana.MeAsUser,
		OrganizationID: "332697157202049",
	})
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range teamsPagesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, team := range page.Teams {
			log.Printf("Page: #%d i: %d team: %#v", pageCount, i, team)
		}
		pageCount += 1
	}
}

func Example_client_AddUserToTeam() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	confirmation, err := client.AddUserToTeam(&asana.TeamRequest{
		UserID: "emm.odeke@gmail.com",
		TeamID: "331783765164429",
	})

	if err != nil {
		log.Fatalf("err adding myself to the team: %v", err)
	}

	log.Printf("confirmation: %#v\n", confirmation)
}

func Example_client_RemoveUserFromTeam() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	err = client.RemoveUserFromTeam(&asana.TeamRequest{
		UserID: "emm.odeke@gmail.com",
		TeamID: "331783765164429",
	})

	if err != nil {
		log.Fatalf("failed to remove user from team, err: %v", err)
	} else {
		log.Printf("successfully removed the user from the team")
	}
}

func Example_client_ListAllUsersInTeam() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	usersPagesChan, _, err := client.ListAllUsersInTeam("331783765164429")
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range usersPagesChan {
		if err := page.Err; err != nil {
			log.Printf("Page: #%d err: %v", pageCount, err)
			continue
		}

		for i, user := range page.Users {
			log.Printf("Page: #%d i: %d team: %#v", pageCount, i, user)
		}
		pageCount += 1
	}
}

func Example_client_FindAttachmentByID() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	foundAttachment, err := client.FindAttachmentByID("338179717217493")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found attachment: %#v\n", foundAttachment)
}

func Example_client_UploadAttachment() {
	imageR, err := os.Open("./testdata/messengerQR.png")
	if err != nil {
		log.Fatal(err)
	}
	defer imageR.Close()

	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	respAttachment, err := client.UploadAttachment(&asana.AttachmentUpload{
		TaskID: "331727965981099",
		Name:   "messenger QR code",
		Body:   imageR,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response attachment: %#v\n", respAttachment)
}

func Example_client_ListAllAttachmentsForTask() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	attachmentsPage, err := client.ListAllAttachmentsForTask("331727965981099")
	if err != nil {
		log.Fatal(err)
	}

	for i, task := range attachmentsPage.Attachments {
		fmt.Printf("Task #%d: %#v\n\n", i, task)
	}
}

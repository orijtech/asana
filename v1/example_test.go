package asana_test

import (
	"fmt"
	"log"

	"github.com/odeke-em/asana/v1"
)

func Example_client_CreateTask() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	setupServers, err := client.CreateTask(&asana.TaskRequest{
		Assignee:  "emm.odeke@gmail.com",
		Notes:     "Please ensure to setup the servers, then ping our group",
		Name:      "server setup",
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
		Workspace: "331783765164429",
	})
	if err != nil {
		log.Fatal(err)
	}

	pageCount := 0
	for page := range taskPagesChan {
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

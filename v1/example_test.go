package asana_test

import (
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

func Example_client_ListTasks() {
}

func Example_client_ListTasksForUser() {
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

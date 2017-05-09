# asana
Asana API client implemented in Go

## Requirements:
* Personal Authentication Token set in your environment as
`ASANA_PERSONAL_ACCESS_TOKEN`
or you could pass in a key when initializing a client.

* Currently their API only has v1 so that's the client we'll use.

## Example creating a task
```go
import (
	"log"

	"github.com/odeke-em/asana/v1"
)

func main() {
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
```

## List all your workspaces
```go
func main() {
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
```

## Find an attachment by id
```go
func main() {
	client, err := asana.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	foundAttachment, err := client.FindAttachmentByID("5678")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found attachment: %#v\n", foundAttachment)
}
```

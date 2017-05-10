# asana
Asana API client implemented in Go

## Requirements:
* Personal Authentication Token set in your environment as
`ASANA_PERSONAL_ACCESS_TOKEN`
or you could pass in a key when initializing a client.

* Currently their API only has v1 so that's the client we'll use.

## Preamble:
```go
import (
	"fmt"
	"log"
	"os"

	"github.com/odeke-em/asana/v1"
)
```

## Example creating a task
```go
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

## Upload an attachment to a task
```go
func main() {
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
```

## List all attachments for a task
```go
func main() {
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
```

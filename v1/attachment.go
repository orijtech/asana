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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/orijtech/otils"

	"github.com/odeke-em/go-uuid"
)

type Attachment struct {
	ID int64 `json:"id,omitempty"`

	CreatedAt   *otils.NullableTime  `json:"created_at,omitempty"`
	DownloadURL otils.NullableString `json:"download_url"`

	// Host is a read-only value.
	// Valid values are asana, dropbox, gdrive and box.
	Host otils.NullableString `json:"host"`

	Name otils.NullableString `json:"name"`

	// Parent contains the information of the
	// task that this attachment is attached to.
	Parent *NamedAndIDdEntity `json:"parent,omitempty"`

	ViewURL otils.NullableString `json:"view_url,omitempty"`
}

var (
	errEmptyAttachmentID = errors.New("expecting a non-empty attachmentID")
	errNoAttachment      = errors.New("no attachment was received")
)

func (c *Client) FindAttachmentByID(attachmentID string) (*Attachment, error) {
	attachmentID = strings.TrimSpace(attachmentID)
	if attachmentID == "" {
		return nil, errEmptyAttachmentID
	}
	fullURL := fmt.Sprintf("%s/attachments/%s", baseURL, attachmentID)
	req, _ := http.NewRequest("GET", fullURL, nil)
	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}
	return parseOutAttachmentFromData(slurp)
}

type AttachmentWrap struct {
	Attachment *Attachment `json:"data"`
}

func parseOutAttachmentFromData(blob []byte) (*Attachment, error) {
	aWrap := new(AttachmentWrap)
	if err := json.Unmarshal(blob, aWrap); err != nil {
		return nil, err
	}
	if aWrap.Attachment != nil {
		return aWrap.Attachment, nil
	}

	return nil, errNoAttachment
}

type AttachmentUpload struct {
	Body   io.Reader `json:"-"`
	TaskID string    `json:"task_id"`
	Name   string    `json:"name"`
}

func (au *AttachmentUpload) nonBlankFilename() string {
	if au.Name != "" {
		return au.Name
	}
	return uuid.NewRandom().String()
}

var errNilBody = errors.New("expecting a non-nil body")

func (au *AttachmentUpload) Validate() error {
	if au == nil || au.Body == nil {
		return errNilBody
	}
	if strings.TrimSpace(au.TaskID) == "" {
		return errEmptyTaskID
	}
	return nil
}

// UploadAtatchment uploads an attachment to a specific task.
// Its fields: TaskID and Body must be set otherwise it will return an error.
func (c *Client) UploadAttachment(au *AttachmentUpload) (*Attachment, error) {
	if err := au.Validate(); err != nil {
		return nil, err
	}

	// Step 1. Try to determine the contentType.
	contentType, body, err := fDetectContentType(au.Body)
	if err != nil {
		return nil, err
	}

	// Step 2:
	// Initiate and then make the upload.
	prc, pwc := io.Pipe()
	mpartW := multipart.NewWriter(pwc)
	go func() {
		defer func() {
			_ = mpartW.Close()
			_ = pwc.Close()
		}()

		formFile, err := mpartW.CreateFormFile("file", au.nonBlankFilename())
		if err != nil {
			return
		}
		_, _ = io.Copy(formFile, body)

		writeStringField(mpartW, "Content-Type", contentType)
		writeStringField(mpartW, "name", au.Name)
	}()

	fullURL := fmt.Sprintf("%s/tasks/%s/attachments", baseURL, au.TaskID)
	req, err := http.NewRequest("POST", fullURL, prc)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mpartW.FormDataContentType())
	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}
	return parseOutAttachmentFromData(slurp)
}

type AttachmentsPage struct {
	Attachments []*Attachment `json:"data"`
}

// ListAllAttachmentsForTask retrieves all the attachments for the taskID provided.
func (c *Client) ListAllAttachmentsForTask(taskID string) (*AttachmentsPage, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errEmptyTaskID
	}
	fullURL := fmt.Sprintf("%s/tasks/%s/attachments", baseURL, taskID)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	slurp, _, err := c.doAuthReqThenSlurpBody(req)
	if err != nil {
		return nil, err
	}

	apage := new(AttachmentsPage)
	if err := json.Unmarshal(slurp, apage); err != nil {
		return nil, err
	}
	return apage, nil
}

func writeStringField(w *multipart.Writer, key, value string) {
	fw, err := w.CreateFormField(key)
	if err == nil {
		_, _ = io.WriteString(fw, value)
	}
}

func fDetectContentType(r io.Reader) (string, io.Reader, error) {
	if r == nil {
		return "", nil, errNilBody
	}

	seeker, seekable := r.(io.Seeker)
	sniffBuf := make([]byte, 512)
	n, err := io.ReadAtLeast(r, sniffBuf, 1)
	if err != nil {
		return "", nil, err
	}

	contentType := http.DetectContentType(sniffBuf)
	needsRepad := !seekable
	if seekable {
		if _, err = seeker.Seek(int64(-n), io.SeekCurrent); err != nil {
			// Since we failed to rewind it, mark it as needing repad
			needsRepad = true
		}
	}

	if needsRepad {
		r = io.MultiReader(bytes.NewReader(sniffBuf), r)
	}

	return contentType, r, nil
}

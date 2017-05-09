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

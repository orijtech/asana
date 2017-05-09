package asana_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/orijtech/asana/v1"
)

func TestFindAttachmentByID(t *testing.T) {
	client, err := asana.NewClient(paToken1)
	if err != nil {
		t.Fatalf("initializing the client: %v", err)
	}

	tests := [...]struct {
		attachmentID string
		wantErr      bool
		want         *asana.Attachment
	}{
		0: {
			attachmentID: attachmentID1,
			want:         attachmentFromFile(attachmentID1),
		},
		1: {
			attachmentID: "",
			wantErr:      true,
		},
		2: {
			attachmentID: "  ",
			wantErr:      true,
		},
	}

	client.SetHTTPRoundTripper(&backend{route: findAttachmentByIDRoute})

	for i, tt := range tests {
		attachment, err := client.FindAttachmentByID(tt.attachmentID)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: wanted non-nil error")
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: got err: %v", i, err)
			continue
		}

		gotBlob := jsonMarshal(attachment)
		wantBlob := jsonMarshal(tt.want)
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\ngotBytes:  %s\nwantBytes: %s", i, gotBlob, wantBlob)
		}
	}
}

const (
	paToken1 = "pa-token-1"

	attachmentID1 = "5678"

	findAttachmentByIDRoute = "find-attachment-by-id"
)

var authorizedTokens = map[string]bool{
	paToken1: true,
}

func authorizedToken(token string) bool {
	_, authorized := authorizedTokens[token]
	return authorized
}

func attachmentPath(attachmentID string) string {
	return fmt.Sprintf("./testdata/attachment-%s.json", attachmentID)
}

func attachmentResponsePath(attachmentID string) string {
	return fmt.Sprintf("./testdata/attachment-response-%s.json", attachmentID)
}

func jsonDeserializeFromFile(path string, save interface{}) error {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, save); err != nil {
		return err
	}
	return nil
}

func attachmentFromFile(attachmentID string) *asana.Attachment {
	path := attachmentPath(attachmentID)
	recv := new(asana.Attachment)
	if err := jsonDeserializeFromFile(path, recv); err != nil {
		return nil
	}
	return recv
}

func jsonMarshal(v interface{}) []byte {
	blob, _ := json.Marshal(v)
	return blob
}

type backend struct {
	route string
}

var _ http.RoundTripper = (*backend)(nil)

func (b *backend) RoundTrip(req *http.Request) (*http.Response, error) {
	switch b.route {
	case findAttachmentByIDRoute:
		return b.findAttachmentByIDRoundTrip(req)
	default:
		return unknownRouteResp, nil
	}
}

func makeResp(status string, code int, body io.ReadCloser) *http.Response {
	resp := &http.Response{
		Status:     status,
		Body:       body,
		Header:     make(http.Header),
		StatusCode: code,
	}

	return resp
}

var (
	unknownRouteResp            = makeResp("unknown route", http.StatusNotFound, nil)
	invalidAuthResp             = makeResp("invalid authentication, make sure to pass \"Bearer <PA_TOKEN>\" in your headers", http.StatusBadRequest, nil)
	unauthorizedBearerTokenResp = makeResp("unauthorized bearer token", http.StatusUnauthorized, nil)
)

func (b *backend) checkAuthorization(req *http.Request, wantMethod string) (*http.Response, error) {
	if got, want := req.Method, wantMethod; got != want {
		return makeResp(fmt.Sprintf("only accepting %q got %q", want, got), http.StatusMethodNotAllowed, nil), nil
	}

	bearerAndPAToken := strings.TrimSpace(req.Header.Get("Authorization"))
	if bearerAndPAToken == "" {
		return invalidAuthResp, nil
	}

	splits := strings.Split(bearerAndPAToken, "Bearer")
	if len(splits) < 2 {
		return invalidAuthResp, nil
	}
	paToken := strings.TrimSpace(splits[len(splits)-1])
	if paToken == "" {
		return invalidAuthResp, nil
	}
	if !authorizedToken(paToken) {
		return unauthorizedBearerTokenResp, nil
	}

	// No fault found, good to go
	return nil, nil
}

func (b *backend) findAttachmentByIDRoundTrip(req *http.Request) (*http.Response, error) {
	if badAuthResp, err := b.checkAuthorization(req, "GET"); err != nil || badAuthResp != nil {
		return badAuthResp, err
	}

	urlPath := strings.Trim(req.URL.Path, "/")
	splits := strings.Split(urlPath, "/")
	if len(splits) < 2 {
		return makeResp("expecting the attachment id", http.StatusBadRequest, nil), nil
	}

	// Last segment of the path
	attachmentID := splits[len(splits)-1]
	diskPath := attachmentResponsePath(attachmentID)
	return makeRespFromFile(diskPath)
}

func makeRespFromFile(path string) (*http.Response, error) {
	f, err := os.Open(path)
	if err != nil {
		return makeResp(err.Error(), http.StatusInternalServerError, nil), nil
	}
	return makeResp("200 OK", http.StatusOK, f), nil
}

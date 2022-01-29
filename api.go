package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/davidmz/mustbe"
)

type UserInfoResponse struct {
	Users struct {
		UserName string `json:"username"`
	} `json:"users"`
}

type TokenInfoResponse struct {
	Token struct {
		Scopes []string `json:"scopes"`
	} `json:"token"`
}

type PlainAttachment struct {
	ID        string `json:"id"`
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	PostID    string `json:"postId"`
}

type Attachment struct {
	PlainAttachment
	CreatedAt time.Time
}

func (a *Attachment) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	if err := json.Unmarshal(data, &a.PlainAttachment); err != nil {
		return err
	}

	ms, _ := strconv.ParseInt(a.PlainAttachment.CreatedAt, 10, 64)
	a.CreatedAt = time.Unix(0, ms*int64(time.Millisecond))
	return nil
}

type AttachmentsListResponse struct {
	Attachments []Attachment `json:"attachments"`
	HasMore     bool         `json:"hasMore"`
}

type ErrorResponse struct {
	Err string `json:"err"`
}

func checkToken() (username string, outErr error) {
	defer mustbe.CatchedAsAnnotated(&outErr, "checking token: %w")

	tokenInfo := new(TokenInfoResponse)
	mustbe.OK(performGetRequest("/v2/app-tokens/current", tokenInfo))
	ok := false
	for _, scope := range tokenInfo.Token.Scopes {
		if scope == "read-my-files" {
			ok = true
			break
		}
	}
	if !ok {
		mustbe.Thrown(fmt.Errorf("missing required permission: %s", "read-my-files"))
	}

	userInfo := new(UserInfoResponse)
	mustbe.OK(performGetRequest("/v1/users/me", userInfo))

	username = userInfo.Users.UserName
	return
}

func getAttachmentsList(page int) (*AttachmentsListResponse, error) {
	attList := new(AttachmentsListResponse)
	err := performGetRequest("/v2/attachments/my?page="+strconv.Itoa(page), attList)
	if err != nil {
		return nil, err
	}
	return attList, nil
}

func performGetRequest(path string, dest interface{}) error {
	pathURL, err := url.Parse(path)
	if err != nil {
		return err
	}
	fullURL := config.apiRootURL.ResolveReference(pathURL)
	req, err := http.NewRequest("GET", fullURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		errResp := new(ErrorResponse)
		if err := json.Unmarshal(body, errResp); err != nil {
			return fmt.Errorf("cannot parse error response: %w", err)
		}
		return fmt.Errorf("API error: %s", errResp.Err)
	}

	if dest != nil {
		if err := json.Unmarshal(body, dest); err != nil {
			return fmt.Errorf("cannot parse response: %w", err)
		}
	}

	return nil
}

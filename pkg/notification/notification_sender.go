package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	http "net/http"
	"strconv"

	"github.com/murilopereirame/dyndns/pkg/client"
)

type NotificationSender struct {
	Url        string
	Secret     string
	AuthHeader string
	HttpClient client.HTTPClient
}

func (s *NotificationSender) Send(payload *NotificationPayload) (bool, error) {
	values := map[string]string{
		"title":    payload.Title,
		"body":     payload.Body,
		"priority": string(payload.Priority),
	}

	requestBody, err := json.Marshal(values)

	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", s.Url, bytes.NewBuffer(requestBody))

	if err != nil {
		return false, err
	}

	req.Header.Add(s.AuthHeader, s.Secret)
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)

	if err != nil {
		return false, err
	}

	if resp.Body == nil {
		return false, errors.New("received empty body while retrieving keys")
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return false, errors.New("request returned non-success code: " + strconv.FormatInt(int64(resp.StatusCode), 10))
	}

	return true, nil
}

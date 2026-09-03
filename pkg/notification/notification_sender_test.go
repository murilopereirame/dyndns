package notification

import (
	"io"
	"net/http"
	"testing"

	"github.com/murilopereirame/dyndns/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSendNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpClient := mock.NewMockHTTPClient(ctrl)

	httpClient.EXPECT().Do(gomock.Cond(func(req *http.Request) bool {
		expectedBody := "{\"body\":\"body\",\"priority\":\"low\",\"title\":\"title\"}"
		bodyBytes, _ := io.ReadAll(req.Body)
		return req.URL.String() == "https://foo.bar" && req.Header.Get("x-api-key") == "secret" && string(bodyBytes) == expectedBody
	})).Return(&http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
	}, nil).Times(1)

	notificationPayload := NotificationPayload{
		Title:    "title",
		Body:     "body",
		Priority: LowPriority,
	}

	sender := NotificationSender{
		Url:        "https://foo.bar",
		Secret:     "secret",
		AuthHeader: "x-api-key",
		HttpClient: httpClient,
	}

	result, error := sender.Send(&notificationPayload)

	assert.NoError(t, error)
	assert.True(t, result)
}

func TestNotificationReject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpClient := mock.NewMockHTTPClient(ctrl)

	httpClient.EXPECT().Do(gomock.Cond(func(req *http.Request) bool {
		expectedBody := "{\"body\":\"body\",\"priority\":\"low\",\"title\":\"title\"}"
		bodyBytes, _ := io.ReadAll(req.Body)
		return req.URL.String() == "https://foo.bar" && req.Header.Get("x-api-key") == "secret" && string(bodyBytes) == expectedBody
	})).Return(&http.Response{
		StatusCode: 400,
		Body:       http.NoBody,
	}, nil).Times(1)

	notificationPayload := NotificationPayload{
		Title:    "title",
		Body:     "body",
		Priority: LowPriority,
	}

	sender := NotificationSender{
		Url:        "https://foo.bar",
		Secret:     "secret",
		AuthHeader: "x-api-key",
		HttpClient: httpClient,
	}

	result, error := sender.Send(&notificationPayload)

	assert.Error(t, error)
	assert.False(t, result)
}

func TestNotificationErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpClient := mock.NewMockHTTPClient(ctrl)

	httpClient.EXPECT().Do(gomock.Cond(func(req *http.Request) bool {
		expectedBody := "{\"body\":\"body\",\"priority\":\"low\",\"title\":\"title\"}"
		bodyBytes, _ := io.ReadAll(req.Body)
		return req.URL.String() == "https://foo.bar" && req.Header.Get("x-api-key") == "secret" && string(bodyBytes) == expectedBody
	})).Return(nil, assert.AnError).Times(1)

	notificationPayload := NotificationPayload{
		Title:    "title",
		Body:     "body",
		Priority: LowPriority,
	}

	sender := NotificationSender{
		Url:        "https://foo.bar",
		Secret:     "secret",
		AuthHeader: "x-api-key",
		HttpClient: httpClient,
	}

	result, error := sender.Send(&notificationPayload)

	assert.ErrorIs(t, assert.AnError, error)
	assert.False(t, result)
}

func TestNotificationNilBodyErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpClient := mock.NewMockHTTPClient(ctrl)

	httpClient.EXPECT().Do(gomock.Cond(func(req *http.Request) bool {
		expectedBody := "{\"body\":\"body\",\"priority\":\"low\",\"title\":\"title\"}"
		bodyBytes, _ := io.ReadAll(req.Body)
		return req.URL.String() == "https://foo.bar" && req.Header.Get("x-api-key") == "secret" && string(bodyBytes) == expectedBody
	})).Return(&http.Response{
		StatusCode: 400,
		Body:       nil,
	}, nil).Times(1)

	notificationPayload := NotificationPayload{
		Title:    "title",
		Body:     "body",
		Priority: LowPriority,
	}

	sender := NotificationSender{
		Url:        "https://foo.bar",
		Secret:     "secret",
		AuthHeader: "x-api-key",
		HttpClient: httpClient,
	}

	result, error := sender.Send(&notificationPayload)

	assert.ErrorContains(t, error, "received empty body while retrieving keys")
	assert.False(t, result)
}

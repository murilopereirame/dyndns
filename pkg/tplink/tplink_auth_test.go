package tplink

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"testing/iotest"

	"github.com/murilopereirame/dyndns/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRetrieveAuthKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().MakeUrl(string(AUTH_KEYS)).Return("https://foo.bar/cgi-bin/luci/;stok=/login?form=keys").Times(1)
	clientMock.EXPECT().MakeUrl("webpages/index.html").Return("https://foo.bar/webpages/index.html").Times(1)
	clientMock.EXPECT().Do(gomock.Cond(func(req *http.Request) bool {
		if err := req.ParseForm(); err != nil {
			return false
		}

		actualBody := req.Form.Encode()
		expectedBody := "form=keys&operation=read"

		originMatches := req.Header.Get("Origin") == "https://foo.bar"
		refererMatches := req.Header.Get("Referer") == "https://foo.bar/webpages/index.html"
		contentTypeMatches := req.Header.Get("Content-Type") == "application/x-www-form-urlencoded"
		acceptMatches := req.Header.Get("Accept") == "application/json, text/plain, */*"
		contentLenMatches := req.ContentLength == int64(len("operation=read"))

		return actualBody == expectedBody &&
			originMatches && refererMatches &&
			contentTypeMatches && acceptMatches &&
			contentLenMatches
	})).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true, "data": { "password": ["00", "00"] }}`)),
	}, nil).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}
	result, err := service.RetrieveAuthKeys()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0}, result.Exponent)
	assert.Equal(t, []byte{0}, result.Modulus)
}

func TestRetrieveAuthKeysWithInvalidResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().MakeUrl(string(AUTH_KEYS)).Return("https://foo.bar/cgi-bin/luci/;stok=/login?form=keys").Times(1)
	clientMock.EXPECT().MakeUrl("webpages/index.html").Return("https://foo.bar/webpages/index.html").Times(1)
	clientMock.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true, "data": [ "password": ["00", "00"]  ]}`)),
	}, nil).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}
	result, err := service.RetrieveAuthKeys()

	assert.ErrorContains(t, err, "received invalid response body")
	assert.Nil(t, result)
}

func TestRetrieveAuthKeysWithInvalidModulusHex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().MakeUrl(string(AUTH_KEYS)).Return("https://foo.bar/cgi-bin/luci/;stok=/login?form=keys").Times(1)
	clientMock.EXPECT().MakeUrl("webpages/index.html").Return("https://foo.bar/webpages/index.html").Times(1)
	clientMock.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true, "data": { "password": ["0", "00"]  }}`)),
	}, nil).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}
	result, err := service.RetrieveAuthKeys()

	assert.ErrorContains(t, err, "encoding/hex: odd length hex string")
	assert.Nil(t, result)
}

func TestRetrieveAuthKeysWithInvalidExponentHex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().MakeUrl(string(AUTH_KEYS)).Return("https://foo.bar/cgi-bin/luci/;stok=/login?form=keys").Times(1)
	clientMock.EXPECT().MakeUrl("webpages/index.html").Return("https://foo.bar/webpages/index.html").Times(1)
	clientMock.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true, "data": { "password": ["00", "0"]  }}`)),
	}, nil).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}
	result, err := service.RetrieveAuthKeys()

	assert.ErrorContains(t, err, "encoding/hex: odd length hex string")
	assert.Nil(t, result)
}

func TestRetrieveAuthKeysWithRequestError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().MakeUrl(string(AUTH_KEYS)).Return("https://foo.bar/cgi-bin/luci/;stok=/login?form=keys").Times(1)
	clientMock.EXPECT().MakeUrl("webpages/index.html").Return("https://foo.bar/webpages/index.html").Times(1)
	clientMock.EXPECT().Do(gomock.Any()).Return(nil, assert.AnError).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}
	result, err := service.RetrieveAuthKeys()
	assert.ErrorIs(t, assert.AnError, err)
	assert.Nil(t, result)
}

func TestRetrieveAuthKeysWithNilBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().MakeUrl(string(AUTH_KEYS)).Return("https://foo.bar/cgi-bin/luci/;stok=/login?form=keys").Times(1)
	clientMock.EXPECT().MakeUrl("webpages/index.html").Return("https://foo.bar/webpages/index.html").Times(1)
	clientMock.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       nil,
	}, nil).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}
	result, err := service.RetrieveAuthKeys()
	assert.ErrorContains(t, err, "received empty body while retrieving keys")
	assert.Nil(t, result)
}

func TestRetrieveAuthKeysWithInvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	errToThrow := errors.New("boom")
	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().MakeUrl(string(AUTH_KEYS)).Return("https://foo.bar/cgi-bin/luci/;stok=/login?form=keys").Times(1)
	clientMock.EXPECT().MakeUrl("webpages/index.html").Return("https://foo.bar/webpages/index.html").Times(1)
	clientMock.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(iotest.ErrReader(errToThrow)),
	}, nil).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}
	result, err := service.RetrieveAuthKeys()
	assert.ErrorIs(t, errToThrow, err)
	assert.Nil(t, result)
}

/*func TestRetrieveRSALoginSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mock.NewMockHTTPClient(ctrl)
	clientMock.EXPECT().Do(func(req *http.Response) {

	}).Return(&http.Response{}, nil).Times(1)

	service := KeyringService{
		middleware: &TPLinkMiddleware{
			httpClient: clientMock,
		},
	}

	service.RetrieveRSALoginSettings()
}*/

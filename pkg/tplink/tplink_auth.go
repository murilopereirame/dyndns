package tplink

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/murilopereirame/dyndns/internal/crypt"
)

type TPLinkPasswordKeys struct {
	Modulus  []byte
	Exponent []byte
}

type RSAAuthSettings struct {
	Key TPLinkPasswordKeys
	Seq int64
}

type AESEncryptionSettings struct {
	Key []byte
	IV  []byte
}

type AuthIdentity struct {
	Stok   string
	Cookie *http.Cookie
}

type TPLinkRouterAuth struct {
	RSAAuthKeys TPLinkPasswordKeys
	AESSettings AESEncryptionSettings
	RSASettings RSAAuthSettings
}

type KeyResponse struct {
	Key []string
	Seq int64
}

type AuthResponse struct {
	Password []string
}

type FormKeysResponse struct {
	Success bool
	Data    KeyResponse
}

type FormAuthResponse struct {
	Success bool
	Data    AuthResponse
}

type KeyringService struct {
	middleware *TPLinkMiddleware
}

func (k *KeyringService) RetrieveAuthKeys() (*TPLinkPasswordKeys, error) {
	req, err := k.middleware.BuildRequest(AUTH_KEYS, "POST")

	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("operation", string(READ))

	encoded := form.Encode()

	req.Body = io.NopCloser(strings.NewReader(encoded))
	req.ContentLength = int64(len(encoded))

	resp, err := k.middleware.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.Body == nil {
		return nil, errors.New("received empty body while retrieving keys")
	}

	defer resp.Body.Close()

	passwordKeys := &TPLinkPasswordKeys{}

	formAuthResponse := &FormAuthResponse{}
	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	isResponseValid := json.Valid(responseBytes)

	if !isResponseValid {
		return nil, errors.New("received invalid response body")
	}

	err = json.Unmarshal(responseBytes, &formAuthResponse)

	if err != nil {
		return nil, err
	}

	modulus, errModulus := hex.DecodeString(formAuthResponse.Data.Password[0])
	exponent, errExponent := hex.DecodeString(formAuthResponse.Data.Password[1])

	if errModulus != nil {
		return nil, errModulus
	}

	if errExponent != nil {
		return nil, errExponent
	}

	passwordKeys.Exponent = exponent
	passwordKeys.Modulus = modulus

	return passwordKeys, nil
}

func (k *KeyringService) RetrieveRSALoginSettings() (*RSAAuthSettings, error) {
	req, err := k.middleware.BuildRequest(RSA_SETTINGS, "POST")

	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("operation", string(READ))

	encoded := form.Encode()

	req.Body = io.NopCloser(strings.NewReader(encoded))
	req.ContentLength = int64(len(encoded))

	resp, err := k.middleware.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	passwordKeys := &TPLinkPasswordKeys{}
	rsaAuthSettings := &RSAAuthSettings{}

	formKeysResponse := &FormKeysResponse{}
	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(responseBytes, &formKeysResponse)

	if err != nil {
		return nil, err
	}

	modulus, errModulus := hex.DecodeString(formKeysResponse.Data.Key[0])
	exponent, errExponent := hex.DecodeString(formKeysResponse.Data.Key[1])

	if errModulus != nil {
		return nil, errModulus
	}

	if errExponent != nil {
		return nil, errExponent
	}

	passwordKeys.Exponent = exponent
	passwordKeys.Modulus = modulus
	rsaAuthSettings.Key = *passwordKeys
	rsaAuthSettings.Seq = formKeysResponse.Data.Seq

	return rsaAuthSettings, nil
}

func (k *KeyringService) GenerateAESEncSettings() (*AESEncryptionSettings, error) {
	key, iv, err := crypt.GenerateAESEncryptionSettings()

	if err != nil {
		return nil, err
	}

	return &AESEncryptionSettings{
		Key: key,
		IV:  iv,
	}, nil
}

func (k *KeyringService) NewRouterAuth() (*TPLinkRouterAuth, error) {
	aes, err := k.GenerateAESEncSettings()

	if err != nil {
		return nil, err
	}

	rsaAuth, err := k.RetrieveAuthKeys()

	if err != nil {
		return nil, err
	}

	rsaLogin, err := k.RetrieveRSALoginSettings()

	if err != nil {
		return nil, err
	}

	return &TPLinkRouterAuth{
		RSAAuthKeys: *rsaAuth,
		AESSettings: *aes,
		RSASettings: *rsaLogin,
	}, nil
}

package tplink

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/murilopereirame/dyndns/internal/crypt"
	"github.com/murilopereirame/dyndns/internal/str"
	"github.com/murilopereirame/dyndns/pkg/client"
)

type NetworkProtocol string

const (
	TCP NetworkProtocol = NetworkProtocol("TCP")
	UDP NetworkProtocol = NetworkProtocol("UDP")
	ALL NetworkProtocol = NetworkProtocol("ALL")
)

type UnauthTPLinkRouter struct {
	WebPanel   string
	Username   string
	Password   string
	RouterAuth *TPLinkRouterAuth
	middleware *TPLinkMiddleware
}

type AuthTPLinkRouter struct {
	WebPanel   string
	RouterAuth *TPLinkRouterAuth
	Identity   *AuthIdentity
	middleware TPLinkMiddleware
}

type LoginResponse struct {
	Data struct {
		Stok string
	}
}

type StatusResponse struct {
	Data struct {
		IPv4 string `json:"wan_ipv4_ipaddr"`
		IPv6 string `json:"wan_ipv6_ip6addr"`
	}
}

type LANClientListResponse struct {
	Data []LANClient
}

type IPv6FirewallRulesListResponse struct {
	Data []IPv6FirewallRule
}

type LANClient struct {
	MacAddress string `json:"mac"`
	Name       string `json:"name"`
	IP         string `json:"ip"`
}

type IPv6FirewallRule struct {
	Key      string
	Port     int             `json:"port,string"`
	Name     string          `json:"name"`
	Enable   string          `json:"enable"`
	Protocol NetworkProtocol `json:"protocol"`
	IP       string          `json:"ip"`
	Type     string          `json:"type"`
}

type IPv6FirewallRuleCRUDResponse struct {
	Success bool
	Data    *IPv6FirewallRule
}

func NewRouter(endpoint, username, password string, httpClient client.HTTPClient) (*UnauthTPLinkRouter, error) {
	middleware := TPLinkMiddleware{
		httpClient: httpClient,
	}
	keyService := KeyringService{
		middleware: &middleware,
	}
	routerAuth, err := keyService.NewRouterAuth()

	if err != nil {
		return nil, err
	}

	return &UnauthTPLinkRouter{
		WebPanel:   endpoint,
		Username:   username,
		Password:   password,
		RouterAuth: routerAuth,
		middleware: &middleware,
	}, nil
}

func (tp *UnauthTPLinkRouter) Authenticate() (*AuthTPLinkRouter, error) {
	credentialHash := crypt.Hash([]byte(tp.Username + tp.Password))
	hexHash := hex.EncodeToString(credentialHash)

	encryptedPassword, err := crypt.EncryptPKCS1v15(tp.RouterAuth.RSAAuthKeys.Modulus, tp.RouterAuth.RSAAuthKeys.Exponent, []byte(tp.Password))

	if err != nil {
		return nil, err
	}

	loginBody := fmt.Sprintf("operation=%s&password=%s&confirm=true", LOGIN, encryptedPassword)
	encryptedBodyBytes, err := crypt.EncryptAES128CBCPKCS7(tp.RouterAuth.AESSettings.Key, tp.RouterAuth.AESSettings.IV, []byte(loginBody))

	if err != nil {
		return nil, err
	}

	base64Body := base64.StdEncoding.EncodeToString(encryptedBodyBytes)

	signStr := fmt.Sprintf(
		"k=%s&i=%s&h=%s&s=%d",
		string(tp.RouterAuth.AESSettings.Key),
		string(tp.RouterAuth.AESSettings.IV),
		string(hexHash),
		tp.RouterAuth.RSASettings.Seq+int64(len(base64Body)),
	)

	signStrChunks := str.SplitStringInChunks(signStr, 53)

	var encryptedChunks []string
	for _, strChunk := range signStrChunks {
		encChunk, err := crypt.EncryptRSAOAEP(
			tp.RouterAuth.RSASettings.Key.Modulus,
			tp.RouterAuth.RSASettings.Key.Exponent,
			[]byte(strChunk),
		)

		if err != nil {
			return nil, err
		}

		encryptedChunks = append(encryptedChunks, encChunk)
	}

	req, err := tp.middleware.BuildRequest(LOGIN_PATH, "POST")

	if err != nil {
		return nil, err
	}

	// The firmware rejects the login with a bodyless 403 unless "sign"
	// precedes "data", so the body can't go through url.Values.Encode(),
	// which sorts fields alphabetically.
	encoded := fmt.Sprintf("sign=%s&data=%s", strings.Join(encryptedChunks, ""), url.QueryEscape(base64Body))

	req.Body = io.NopCloser(strings.NewReader(encoded))
	req.ContentLength = int64(len(encoded))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsedResponse := &RouterResponse{}
	err = json.Unmarshal(bodyBytes, &parsedResponse)

	if err != nil {
		fmt.Println("Failed to parse json response " + err.Error())
		return nil, nil
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(parsedResponse.Data)

	if err != nil {
		fmt.Println("Failed to decode bas64 response " + err.Error())
		return nil, err
	}

	decryptedData, err := crypt.DecryptAES128CBCPKCS7(tp.RouterAuth.AESSettings.Key, tp.RouterAuth.AESSettings.IV, decodedBytes)

	if err != nil {
		fmt.Println("Failed to decrypt response " + err.Error())
		return nil, err
	}

	decodedLoginResponse := &LoginResponse{}
	err = json.Unmarshal(decryptedData, &decodedLoginResponse)

	if err != nil {
		fmt.Println("Failed to parse decrypted json response " + err.Error())
		return nil, err
	}

	cookieIndex := slices.IndexFunc(resp.Cookies(), func(cookie *http.Cookie) bool {
		return cookie.Name == SYS_AUTH_COOKIE
	})

	if cookieIndex == -1 {
		fmt.Println("Failed to find session cookie")
		return nil, err
	}

	authIdentity := &AuthIdentity{
		Stok:   decodedLoginResponse.Data.Stok,
		Cookie: resp.Cookies()[cookieIndex],
	}

	authRouter := &AuthTPLinkRouter{
		WebPanel:   tp.WebPanel,
		RouterAuth: tp.RouterAuth,
		Identity:   authIdentity,
	}

	return authRouter, nil
}

func (tp *AuthTPLinkRouter) GetIPv4Address() (string, error) {
	jsonResponse, err := tp.Request(string(ALL_STATUS), READ, nil)

	if err != nil {
		return "", err
	}

	parsedResponse := &StatusResponse{}
	err = json.Unmarshal([]byte(jsonResponse), &parsedResponse)

	if err != nil {
		return "", err
	}

	return parsedResponse.Data.IPv4, nil
}

func (tp *AuthTPLinkRouter) GetIPv6Address() (string, error) {
	jsonResponse, err := tp.Request(string(ALL_STATUS), READ, nil)

	if err != nil {
		return "", err
	}

	parsedResponse := &StatusResponse{}
	err = json.Unmarshal([]byte(jsonResponse), &parsedResponse)

	if err != nil {
		return "", err
	}

	return strings.Split(parsedResponse.Data.IPv6, "/")[0], nil
}

func (tp *AuthTPLinkRouter) Request(endpoint string, operation RouterOperations, partialBody map[string]any) (string, error) {
	path := strings.ReplaceAll(endpoint, ":stok", tp.Identity.Stok)
	req, err := tp.middleware.BuildRequest(RouterEndpoint(path), "POST")

	if err != nil {
		return "", err
	}

	req.AddCookie(tp.Identity.Cookie)

	body := fmt.Sprintf("operation=%s", operation)

	for key, value := range partialBody {
		strValue := fmt.Sprintf("%v", value)
		body = fmt.Sprintf("%s&%s=%s", body, key, url.QueryEscape(strValue))
	}

	slog.Debug("Ongoing request to router", "path", path, "body", body)

	encryptedBodyBytes, err := crypt.EncryptAES128CBCPKCS7(tp.RouterAuth.AESSettings.Key, tp.RouterAuth.AESSettings.IV, []byte(body))

	if err != nil {
		return "", err
	}

	encodedBody := base64.StdEncoding.EncodeToString(encryptedBodyBytes)
	rawHashBody := crypt.Hash([]byte(encodedBody))
	hexHash := hex.EncodeToString(rawHashBody)

	signStr := fmt.Sprintf("h=%s&s=%d", hexHash, tp.RouterAuth.RSASettings.Seq+int64(len(encodedBody)))
	signStrChunks := str.SplitStringInChunks(signStr, 53)

	hmacKey := fmt.Sprintf("k=%s&i=%s", string(tp.RouterAuth.AESSettings.Key), string(tp.RouterAuth.AESSettings.IV))
	var authenticatedChunks []string
	for _, strChunk := range signStrChunks {
		authChunk := crypt.HMACSHA256([]byte(hmacKey), []byte(strChunk))
		hexAuthChunk := hex.EncodeToString(authChunk)
		authenticatedChunks = append(authenticatedChunks, hexAuthChunk)
	}

	encoded := fmt.Sprintf("sign=%s&data=%s", strings.Join(authenticatedChunks, ""), url.QueryEscape(encodedBody))

	req.Body = io.NopCloser(strings.NewReader(encoded))
	req.ContentLength = int64(len(encoded))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Request failed - %s\nPath: %s", resp.Status, path)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	parsedResponse := &RouterResponse{}
	err = json.Unmarshal(bodyBytes, &parsedResponse)

	if err != nil {
		fmt.Println("Failed to parse json response " + err.Error())
		return "", err
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(parsedResponse.Data)

	if err != nil {
		fmt.Println("Failed to decode bas64 response " + err.Error())
		return "", err
	}

	decryptedData, err := crypt.DecryptAES128CBCPKCS7(tp.RouterAuth.AESSettings.Key, tp.RouterAuth.AESSettings.IV, decodedBytes)

	if err != nil {
		fmt.Println("Failed to decrypt response " + err.Error())
		return "", err
	}

	return string(decryptedData), nil
}

func (tp *AuthTPLinkRouter) GetIPv4ClientList() ([]LANClient, error) {
	jsonResponse, err := tp.Request(string(IPV4_CLIENT_LIST), READ, nil)

	if err != nil {
		return nil, err
	}

	parsedResponse := &LANClientListResponse{}
	err = json.Unmarshal([]byte(jsonResponse), &parsedResponse)

	if err != nil {
		return nil, err
	}

	return parsedResponse.Data, nil
}

func (tp *AuthTPLinkRouter) GetIPv6ClientList() ([]LANClient, error) {
	jsonResponse, err := tp.Request(string(IPV6_CLIENT_LIST), READ, nil)

	if err != nil {
		return nil, err
	}

	parsedResponse := &LANClientListResponse{}
	err = json.Unmarshal([]byte(jsonResponse), &parsedResponse)

	if err != nil {
		return nil, err
	}

	return parsedResponse.Data, nil
}

func (tp *AuthTPLinkRouter) FetchIPv6FirewallRules() ([]IPv6FirewallRule, error) {
	jsonResponse, err := tp.Request(string(IPV6_FIREWALL), READ, nil)

	if err != nil {
		return nil, err
	}

	parsedResponse := &IPv6FirewallRulesListResponse{}
	err = json.Unmarshal([]byte(jsonResponse), &parsedResponse)

	if err != nil {
		return nil, err
	}

	return parsedResponse.Data, nil
}

func (tp *AuthTPLinkRouter) AddPortInIPv6Firewall(rule IPv6FirewallRule, onConflict RuleConflictResolution) (bool, error) {
	existingRules, err := tp.FetchIPv6FirewallRules()

	if err != nil {
		return false, err
	}

	ruleWithSameNameIndex := slices.IndexFunc(existingRules, func(r IPv6FirewallRule) bool {
		return r.Name == rule.Name
	})

	ruleWithSameConfigIndex := slices.IndexFunc(existingRules, func(r IPv6FirewallRule) bool {
		hasSameIPAndPort := r.IP == rule.IP && r.Port == rule.Port
		hasSameProtocol := r.Protocol == rule.Protocol

		return hasSameIPAndPort && hasSameProtocol
	})

	ruleIndex := 0
	var existingRule *IPv6FirewallRule
	if ruleWithSameConfigIndex >= 0 {
		existingRule = &existingRules[ruleWithSameConfigIndex]
		ruleIndex = ruleWithSameConfigIndex
		slog.Info("A rule for this IP, Port and Protocol already exists", "ip", rule.IP, "port", rule.Port, "protocol", rule.Protocol, "onConflict", onConflict)
	} else if ruleWithSameNameIndex >= 0 {
		existingRule = &existingRules[ruleWithSameNameIndex]
		ruleIndex = ruleWithSameNameIndex
		slog.Info("A rule with the same name already exists", "name", rule.Name, "onConflict", onConflict)
	}

	operation := INSERT
	if existingRule != nil && onConflict == OVERWRITE {
		operation = UPDATE
	} else if existingRule != nil {
		slog.Warn("Skipping rule overwrite", "ip", rule.IP, "port", rule.Port, "protocol", rule.Protocol, "name", rule.Name, "onConflict", onConflict)
		return false, nil
	}

	newRuleBytes, err := json.Marshal(rule)

	if err != nil {
		return false, err
	}
	encodedNewRule := string(newRuleBytes)

	partialBody := make(map[string]any, 0)
	partialBody["new"] = encodedNewRule

	if operation == UPDATE {
		oldRuleBytes, err := json.Marshal(existingRule)

		if err != nil {
			return false, err
		}

		encodedOldRule := string(oldRuleBytes)
		partialBody["old"] = string(encodedOldRule)
	}

	partialBody["index"] = ruleIndex

	resultJson, err := tp.Request(string(IPV6_FIREWALL), operation, partialBody)

	crudResponse := &IPv6FirewallRuleCRUDResponse{}
	err = json.Unmarshal([]byte(resultJson), &crudResponse)

	if err != nil {
		return false, err
	}

	if !crudResponse.Success {
		errorResponse := &ErrorResponse{}
		err = json.Unmarshal([]byte(resultJson), errorResponse)

		if err != nil {
			return false, tp.middleware.ParseWebPanelError("")
		}

		return false, tp.middleware.ParseWebPanelError(errorResponse.ErrorCode)
	}

	return crudResponse.Success, nil
}

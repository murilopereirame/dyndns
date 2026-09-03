package tplink

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/murilopereirame/dyndns/pkg/client"
)

type RouterOperations string

const (
	READ   RouterOperations = RouterOperations("read")
	INSERT RouterOperations = RouterOperations("insert")
	UPDATE RouterOperations = RouterOperations("update")
	REMOVE RouterOperations = RouterOperations("remove")
	LOGIN  RouterOperations = RouterOperations("login")
)

type RuleConflictResolution string

const (
	IGNORE    RuleConflictResolution = RuleConflictResolution("IGNORE")
	OVERWRITE RuleConflictResolution = RuleConflictResolution("OVERWRITE")
)

type RouterEndpoint string

const (
	AUTH_KEYS        RouterEndpoint = RouterEndpoint("cgi-bin/luci/;stok=/login?form=keys")
	RSA_SETTINGS     RouterEndpoint = RouterEndpoint("cgi-bin/luci/;stok=/login?form=auth")
	LOGIN_PATH       RouterEndpoint = RouterEndpoint("cgi-bin/luci/;stok=/login?form=login")
	ALL_STATUS       RouterEndpoint = RouterEndpoint("cgi-bin/luci/;stok=:stok/admin/status?form=all")
	IPV4_CLIENT_LIST RouterEndpoint = RouterEndpoint("cgi-bin/luci/;stok=:stok/admin/nat?form=client_list")
	IPV6_CLIENT_LIST RouterEndpoint = RouterEndpoint("cgi-bin/luci/;stok=:stok/admin/nat?form=client_list_v6")
	IPV6_FIREWALL    RouterEndpoint = RouterEndpoint("cgi-bin/luci/;stok=:stok/admin/nat?form=fr6")
)

const (
	SYS_AUTH_COOKIE string = "sysauth"
)

type RouterResponse struct {
	Data string
}

type ErrorResponse struct {
	ErrorCode string `json:"errorcode"`
	Success   bool
}

type TPLinkMiddleware struct {
	httpClient client.HTTPClient
}

func (t *TPLinkMiddleware) ParseWebPanelError(code string) error {
	switch code {
	case "00000001":
		return errors.New("Invalid file type.")
	case "00000002":
		return errors.New("Checksum error.")
	case "00000003":
		return errors.New("The file is too large.")
	case "00000004":
		return errors.New("Upload error.")
	case "00000005":
		return errors.New("Reboot error.")
	case "00000006":
		return errors.New("Unknown error.")
	case "00000007":
		return errors.New("The item already exists. Please enter another one.")
	case "00000045":
		return errors.New("Default Gateway and LAN IP address should be in the same subnet.")
	case "00000047":
		return errors.New("The IP address and the router's LAN IP address should be in the same subnet.")
	case "00000051":
		return errors.New("This MAC address already exists. Please enter another one.")
	case "00000056":
		return errors.New("The remote IP address and the current LAN IP address cannot be in the same subnet.")
	case "00000059":
		return errors.New("Invalid Subnet Mask and LAN IP address.")
	case "00000060":
		return errors.New("WAN IP address and LAN IP address cannot be in the same subnet.")
	case "00000062":
		return errors.New("This field is required.")
	case "00000064":
		return errors.New("Failed to block - you are managing the router with this device.")
	case "00000065":
		return errors.New("This item conflicts with existing ones. Please try again.")
	case "00000066":
		return errors.New("The password should be 8 to 63 characters or 64 hexadecimal digits.")
	case "00000077":
		return errors.New("The IP address cannot be the same as the LAN IP address.")
	case "00000080":
		return errors.New("Passwords mismatch. Please try again.")
	case "00000088":
		return errors.New("This operation is not allowed for remote management.")
	case "00000094":
		return errors.New("The VLAN IDs cannot be the same.")
	case "00000095":
		return errors.New("At least one Internet port is required.")
	case "00000107":
		return errors.New("The destination already exists.")
	case "00000110":
		return errors.New("The IP address and LAN IP address should be in the same subnet.")
	case "00000131":
		return errors.New("NTP Server cannot be a loopback address.")
	case "00000150":
		return errors.New("Invalid Subnet Mask and LAN IP address.")
	case "00000250":
		return errors.New("The folder does not exist.")
	case "00000251":
		return errors.New("Please specify a folder for the downloads.")
	case "00000252":
		return errors.New("Download folder error.")
	case "00000253":
		return errors.New("USB device is unplugged!")
	case "00000254":
		return errors.New("No USB device!")
	case "00000256":
		return errors.New("Content must be up to 31 characters.")
	case "00000257":
		return errors.New("Invalid file format.")
	case "00000259":
		return errors.New("Password must be 8-16 alpha characters, numbers, and _.")
	case "00000266":
		return errors.New("Disconnected from the internet. Please check the hardware connection and then try again.")
	case "00000273":
		return errors.New("This download item already exists.")
	case "00004193":
		return errors.New("The password should be 8 to 63 characters or hexadecimal digits.")
	default:
		return errors.New("Unknown error from WebPanel")
	}
}

func (t *TPLinkMiddleware) BuildRequest(path RouterEndpoint, method string) (*http.Request, error) {
	endpoint := t.httpClient.MakeUrl(string(path))
	referer := t.httpClient.MakeUrl("webpages/index.html")

	req, err := http.NewRequest(method, endpoint, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Origin", fmt.Sprintf("%s://%s", req.URL.Scheme, req.URL.Host))
	req.Header.Add("Referer", referer)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json, text/plain, */*")

	return req, nil
}

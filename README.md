# dyndns — Dynamic DNS for Cloudflare and TPLink routers

A tiny Go service which checks periodically for IP changes and update the respective zones in Cloudflare and firewall in TPLink router.

---

## Features
- IPv4 and IPv6 support
- DNS records auto update
- Firewall rules auto management
- Simple JSON config file
- Docker enabled

---

## Quick start (local)
Prereqs: Go (per go.mod), or use Docker below.

1) Copy and edit config
```bash
cp config.example.json my-config.json
# edit my-config.json to match your needs
```

2) Run
```bash
CONFIG_PATH=my-config.json go run .
```

3) Build (optional)
```bash
go build -o dyndns .
CONFIG_PATH=my-config.json ./dyndns
```

4) Test
```bash
go test ./...
```

---

## Docker
Build a local image:
```bash
docker build -t dyndns:local .
```

Run with bind mounts and non-root user:
```bash
docker run --rm \
  -e CONFIG_PATH=/app/config.json \
  -v $(pwd)/my-config.json:/app/config.json:ro \
  -v /path/on/host:/files \
  dyndns:local
```

- PUID/PGID build args are supported to set the container user (defaults 1000/1000). If needed, build like:
```bash
docker build --build-arg PUID=$(id -u) --build-arg PGID=$(id -g) -t dyndns:local .
```

---

## Docker Compose
A compose file is included. Edit the volume paths for your system and config.

Example:
```yaml
services:
  dyndns:
    build:
      context: .
      args:
        - "PUID=${PUID:-1000}"
        - "PGID=${PGID:-1000}"
    container_name: "dyndns"
    network_mode: "host"
    restart: unless-stopped
    environment:
      - CONFIG_PATH=config.json
    volumes:
      - ./config.json:/app/config.json:ro
```

Start it:
```bash
docker compose up -d --build
```

---

## Environment
- CONFIG_PATH: optional; path to the JSON config (default `config.json` in working directory). In Docker Compose we use `/app/config.json`.

---

## Example config

```json
{
  "interval": 60,
  "webhook": {
    "url": "",
  	"secret": "",
  	"authHeader": "",
    "priority": "normal",
  	"enabled": true
  },
  "router": {
   	"endpoint": "",
  	"user": "",
  	"password": "",
  	"ipv4": false,
  	"ipv6": false
  },
  "domain": {
    "domain": "",
    "email": "",
   	"token": "",
   	"zoneId": "",
   	"proxied": false
  },
  "nic": {
   	"interface": "",
  	"ipv4": false,
  	"ipv6": true
  },
  "firewall": {
   	"port": 3000,
  	"name": "React",
  	"protocol": "ALL",
  	"onConflict": "OVERWRITE",
    "enabled": true
  }
}
```

This config runs a check every 60 seconds, checking Cloudflare and the NIC for IP changes. It will prioritize the IP from NIC and ignore the the Router IP. If the IP has changed, the firewall will be updated.

| Field | Type | Description | Required? |
|---|---|---|---|
| `interval` | int64 | Seconds between each IP-change check loop. | **Yes** — always used. |
| `router.endpoint` | string | Base URL of the TP-Link router admin interface. | Conditionally — required if `router.ipv4`, `router.ipv6`, **or `firewall.enabled`** is `true`. |
| `router.user` | string | TP-Link router login username. | Conditionally — same condition as `router.endpoint`. |
| `router.password` | string | TP-Link router login password. | Conditionally — same condition as `router.endpoint`. |
| `router.ipv4` | bool | Use the router as the IPv4 address source. | No — defaults to `false`. |
| `router.ipv6` | bool | Use the router as the IPv6 address source. | No — defaults to `false`. |
| `domain.domain` | string | The DNS record name (e.g. `home.example.com`) to keep updated in Cloudflare. | **Yes** — always used (state is fetched from Cloudflare on startup). |
| `domain.email` | string | Cloudflare API email with DNS edit permission for the zone. | **Yes** |
| `domain.token` | string | Cloudflare API token with DNS edit permission for the zone. | **Yes** |
| `domain.zoneId` | string | Cloudflare Zone ID that owns `domain.domain`. | **Yes** |
| `domain.proxied` | bool | Whether the DNS record is proxied through Cloudflare (orange cloud). | No — defaults to `false`. |
| `nic.interface` | string | Local network interface name (e.g. `eth0`) to read the IP from. | Conditionally — required only if `nic.ipv4` or `nic.ipv6` is `true`. |
| `nic.ipv4` | bool | Use the local NIC as the IPv4 address source. | No — defaults to `false`. |
| `nic.ipv6` | bool | Use the local NIC as the IPv6 address source. | No — defaults to `false`. |
| `firewall.port` | int | Port number for the firewall rule created/updated on the router. | Conditionally — required only if `firewall.enabled` is `true`. |
| `firewall.name` | string | Display name for the firewall rule. | Conditionally — same as above. |
| `firewall.protocol` | string enum: `TCP`, `UDP`, `ALL` | Protocol the firewall rule applies to. | Conditionally — same as above. |
| `firewall.onConflict` | string enum: `IGNORE`, `OVERWRITE` | How to resolve a naming/rule conflict on the router. | Conditionally — same as above. |
| `firewall.enabled` | bool | Enables automatic IPv6 firewall rule management when the IPv6 address changes. | No — defaults to `false`. |
| `webhook.url` | string | Endpoint the notification is POSTed to when a DNS record changes. | Conditionally — required only if `webhook.enabled` is `true`. |
| `webhook.secret` | string | Value sent in the `authHeader` header for authenticating the webhook call. | Conditionally — same as above. |
| `webhook.authHeader` | string | Name of the HTTP header used to carry `webhook.secret`. | Conditionally — same as above. |
| `webhook.priority` | string enum: `low`, `normal`, `high` | Priority tag included in the notification payload. | No — cosmetic; only meaningful if `webhook.enabled` is `true`. |
| `webhook.enabled` | bool | Enables sending a webhook notification after a successful IP/DNS update. | No — defaults to `false`. |


---

## 3rd Party Disclaimer

This project integrates with and interacts with third-party services and hardware, including Cloudflare and TP-Link routers. Cloudflare and TP-Link are independent third parties and are **not affiliated with, sponsored by, endorsed by, or officially associated with this project**.

This project is not an official Cloudflare or TP-Link product. Any references to their products, services, APIs, or trademarks are made solely to describe the functionality and integrations supported by this software.

---

## License

This project is licensed under MIT License. See [LICENSE](LICENSE) for details.

---

Thanks for using **DYNDNS**! If you need help or have ideas to improve it, feel free to raise issues or submit pull requests.

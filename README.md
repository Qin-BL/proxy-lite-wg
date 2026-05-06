# Proxy Lite TLS

`proxy-lite-wg` now runs as a lightweight control plane for `VLESS over WebSocket over TLS`.

It keeps the original goals:

- one shared Linux edge machine
- one client identity per device
- instant revoke / disable for a single client
- simple browser-based management

The transport has changed from WireGuard to:

- `Xray`
- `VLESS`
- `WebSocket`
- TLS terminated by the existing OpenResty / Nginx HTTPS edge

## What It Does

- admin login with session cookie
- user management
- per-device client issuance
- copyable VLESS share links
- QR code export
- disable / enable / delete client
- automatic Xray config regeneration
- automatic Xray container restart after client changes

## Default Topology

```text
Browser / v2rayN / mobile client
        |
        | TLS 443
        v
OpenResty (existing shared domain)
        |
        | WS path proxy
        v
Xray inbound on 127.0.0.1:10000
        |
        v
Internet
```

## Quick Start

1. Copy `.env.example` to `.env`
2. Fill in:
   - `ADMIN_USERNAME`
   - `ADMIN_PASSWORD` or `ADMIN_PASSWORD_HASH`
   - `PUBLIC_HOST`
   - `VLESS_WS_PATH`
3. Run:

```powershell
go run ./cmd/proxy-lite-wg
```

Default local address:

```text
http://127.0.0.1:8080/manage
```

## Deploy

Use the files under `deploy/`.

- `deploy/docker-compose.yml`
- `deploy/.env.example`
- `deploy/openresty/xray-vless-ws.conf.example`

The control plane writes the live Xray config to the shared runtime volume and restarts the `proxy-lite-wg-xray` container whenever clients change.

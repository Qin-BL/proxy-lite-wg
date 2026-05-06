# Deployment

This project now deploys two services:

- `control-plane`: admin UI + SQLite + Xray config generator
- `xray`: VLESS over WebSocket runtime

## Ports

- `127.0.0.1:18080/tcp`: control plane
- `127.0.0.1:10000/tcp`: Xray inbound for OpenResty to proxy to

The public HTTPS entrypoint stays on the existing OpenResty instance that already owns `443/tcp`.

## Deploy Steps

1. Copy `deploy/.env.example` to `deploy/.env`
2. Set:
   - `ADMIN_USERNAME`
   - `ADMIN_PASSWORD` or `ADMIN_PASSWORD_HASH`
   - `PUBLIC_HOST`
   - `VLESS_WS_PATH`
3. Bring the stack up:

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

## OpenResty Snippet

Use `deploy/openresty/xray-vless-ws.conf.example` as the location block mounted into the shared HTTPS domain.

Example upstream path:

```nginx
location = /_ws_change_me {
    proxy_pass http://127.0.0.1:10000;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400;
}
```

## Management UI

Expose the existing `/manage` route only through your shared HTTPS domain path mapping.

The UI now requires admin login before any user or client operations.

## Runtime Apply Flow

Whenever a client is created, enabled, disabled, or deleted:

1. SQLite is updated
2. a fresh Xray config file is written
3. the `proxy-lite-wg-xray` container is restarted through the Docker socket

That makes client revocation take effect immediately.

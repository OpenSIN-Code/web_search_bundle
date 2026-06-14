<!-- Purpose: Production deployment guide for the sin-websearch HTTP API. -->
<!-- Docs: docs/deployment.md.doc.md -->

# Deployment Guide

## Overview

This guide covers production deployment of the `sin-websearch` HTTP API server (`sin-websearch http`). It includes:

- TLS termination with Caddy or nginx
- Running the binary as a systemd service
- Running with Docker
- Enabling bearer token authentication
- Configuring per-IP rate limiting

The API listens on port `8787` by default. Do not expose this port directly on the public internet without TLS and authentication.

## TLS termination with Caddy

[Caddy](https://caddyserver.com/) is the simplest option because it provisions and renews TLS certificates automatically.

```caddyfile
# Caddyfile
api.example.com {
    reverse_proxy localhost:8787
}
```

Run Caddy:

```bash
caddy run --config Caddyfile
```

Caddy obtains a certificate for `api.example.com` from Let's Encrypt (or ZeroSSL) and proxies all traffic to the `sin-websearch` HTTP server on `localhost:8787`.

## TLS termination with nginx

Use nginx as a reverse proxy when you prefer manual certificate management.

```nginx
server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate     /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;

    location / {
        proxy_pass         http://localhost:8787;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
    }
}

server {
    listen 80;
    server_name api.example.com;
    return 301 https://$server_name$request_uri;
}
```

Reload nginx after editing the configuration:

```bash
sudo nginx -t && sudo systemctl reload nginx
```

## systemd service

Run `sin-websearch http` as a system service so it starts automatically and restarts on failure.

```ini
# /etc/systemd/system/sin-websearch.service
[Unit]
Description=sin-websearch HTTP API
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/sin-websearch http
Restart=on-failure
RestartSec=5
User=sin-websearch
Group=sin-websearch
Environment="HOME=/var/lib/sin-websearch"
WorkingDirectory=/var/lib/sin-websearch

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sin-websearch

[Install]
WantedBy=multi-user.target
```

Create the user and data directory, then enable the service:

```bash
sudo useradd -r -m -d /var/lib/sin-websearch -s /usr/sbin/nologin sin-websearch
sudo mkdir -p /var/lib/sin-websearch/.config/sin-websearch
sudo install -m 755 -o sin-websearch -g sin-websearch /path/to/sin-websearch /usr/local/bin/sin-websearch
sudo systemctl daemon-reload
sudo systemctl enable --now sin-websearch
sudo systemctl status sin-websearch
```

Place `sin-websearch.yaml` in `/var/lib/sin-websearch/.config/sin-websearch/` and restrict its permissions to the service user.

## Docker

Run the pre-built image from a GitHub Container Registry (replace `ghcr.io/opensin-code/sin-websearch:latest` with the tag you publish):

```bash
# Create a config directory on the host
mkdir -p "$HOME/.config/sin-websearch"

# Run the container
docker run -d \
  --name sin-websearch \
  -p 127.0.0.1:8787:8787 \
  -v "$HOME/.config/sin-websearch:/etc/sin-websearch" \
  -e SIN_WEBSEARCH_CONFIG=/etc/sin-websearch/sin-websearch.yaml \
  ghcr.io/opensin-code/sin-websearch:latest
```

For a multi-service deployment with Caddy or nginx, use Docker Compose:

```yaml
# docker-compose.yml
services:
  sin-websearch:
    image: ghcr.io/opensin-code/sin-websearch:latest
    restart: unless-stopped
    volumes:
      - ./sin-websearch.yaml:/etc/sin-websearch/sin-websearch.yaml:ro
    environment:
      - SIN_WEBSEARCH_CONFIG=/etc/sin-websearch/sin-websearch.yaml
    networks:
      - sin-websearch

  caddy:
    image: caddy:2
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    networks:
      - sin-websearch

volumes:
  caddy_data:
  caddy_config:

networks:
  sin-websearch:
```

Build your own image if you do not use the published registry image:

```dockerfile
# Dockerfile
FROM golang:1.25 AS builder
WORKDIR /src
COPY . .
RUN go mod download && go build -o /bin/sin-websearch ./cmd/sin-websearch

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/sin-websearch /bin/sin-websearch
EXPOSE 8787
ENTRYPOINT ["/bin/sin-websearch"]
CMD ["http"]
```

## Bearer token authentication

All HTTP endpoints support optional bearer-token authentication when a `token` is configured in `sin-websearch.yaml`:

```yaml
token: "your-long-random-token"
```

Clients must include the token in every request:

```bash
curl -X POST https://api.example.com/api/v1/search \
  -H 'Authorization: Bearer your-long-random-token' \
  -H 'Content-Type: application/json' \
  -d '{"query":"OpenSIN"}'
```

Use a cryptographically random token of at least 32 bytes. Store it in a secrets manager, not in version control.

## Rate limiting

The API applies per-IP token-bucket rate limiting on all mutating and expensive endpoints. Defaults are `10` requests per second with a burst of `20`. Override them in `sin-websearch.yaml`:

```yaml
rate_limit_rps: 10.0      # per-IP requests per second
rate_limit_burst: 20      # per-IP burst capacity
```

When running behind a reverse proxy, the server sees the proxy's IP instead of the client's IP. Make sure the reverse proxy forwards the real client IP (`X-Forwarded-For`) if you need the limiter to apply per-client. For deployments that need more accurate per-client limiting, terminate the proxy header correctly and consider a proxy-aware rate-limiting layer.

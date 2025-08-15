# NSM

Local development platform for clean HTTPS URLs, automatic certificates and DNS routing for dev projects.

## Table of contents
- [Project Structure](#project-structure)
- [Platform Support](#platform-support)
- [Install & Build](#install--build)
- [Usage & Local Testing](#usage--local-testing)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Release](#release)
- [Contact](#contact)

## Project Structure
Top-level layout (trimmed to most relevant paths):
- cmd/
    - nsm/ (CLI server)
    - setup/ (nsm-setup installer)
- examples/
    - go/
    - java/
    - python/
    - react-vite-typescript/
    - rust/
- internal/ (core app, cert, dns, platform, project detection, setup UI)
- pkg/ (logger, utils)
- build/ (output)
- Makefile, go.mod, README.md

Use the example projects in `examples/` to validate behavior across runtimes.

## Platform Support

Current Support Status (v1.0.0)

- ✅ macOS (Full)
    - dnsmasq + resolver integration, mkcert integration, clean HTTPS on port 443 without sudo (macOS flow)
- ⚠️ Linux (Partial)
    - dnsmasq and mkcert supported, manual DNS/systemd-resolved steps and port 443 may require elevated privileges
- ❌ Windows
    - Not supported in v1.0.0 (planned for future releases)

## Install & Build

Prerequisites
- Go (>= 1.21 recommended)
- mkcert, dnsmasq (install instructions differ per platform; Homebrew on macOS recommended)

Build from source
```bash
# From repo root
git clone https://github.com/kashifsb/nsm.git
cd nsm

# Build binaries (Makefile targets used in this repo)
make all setup-build

# or build directly
go build -o build/nsm ./cmd/nsm
go build -o build/nsm-setup ./cmd/setup

# Install locally (optional)
sudo mv build/nsm /usr/local/bin/nsm
sudo mv build/nsm-setup /usr/local/bin/nsm-setup
```

## Usage & Local Testing

1. Setup system integration (installs mkcert CA and configures dnsmasq/resolvers where supported)
```bash
# Interactive installer (macOS recommended for full automation)
nsm-setup install
nsm-setup status
nsm-setup tld list
```

2. Start a project with NSM
```bash
# Example: run nsm for a detected project in current directory
nsm --project-type vite --domain myapp.dev
```

3. Test access (examples)
```bash
# Clean HTTPS (port 443)
curl -k https://myapp.dev

# Fallback (explicit port)
curl -k https://localhost:8443

# DNS resolution
nslookup myapp.dev
dig myapp.dev
```

4. Cleanup after testing
```bash
nsm-setup reset
nsm-setup tld remove test
make clean
```

## Examples

The repository includes runnable sample apps in `examples/`. Each example typically includes a small launcher under its `cmd/` directory or standard run instructions:

- React + Vite (examples/react-vite-typescript)
    ```bash
    cd examples/react-vite-typescript
    npm install
    # start your app (see project README or cmd/)
    # then in another terminal
    nsm --project-type vite --domain react-example.dev
    # Visit: https://react-example.dev
    ```

- Go (examples/go)
    ```bash
    cd examples/go
    go run ./...
    nsm --project-type go --domain go-example.dev
    # Visit: https://go-example.dev
    ```

- Python (examples/python)
    ```bash
    cd examples/python
    python -m venv venv
    source venv/bin/activate
    pip install -r requirements.txt
    python app.py
    nsm --project-type python --domain py-example.dev
    ```

- Java, Rust: each example contains a small app and `cmd/` helpers — follow the per-example README or start scripts and then run `nsm` with an appropriate `--domain`.

If an example provides a `cmd/` helper script, prefer that to start the demo app.

## Troubleshooting

- Port conflicts:
    lsof -i :443
- DNS/DNSMasq:
    - macOS: brew services list | grep dnsmasq
    - Check /etc/resolver/ files
- mkcert:
    mkcert -CAROOT
    mkcert -install
- systemd-resolved conflicts (Linux): manual resolution required until v1.1.x improvements

## Release

A GitHub Actions release workflow (build matrix for darwin/linux + amd64/arm64) is used in this project. Tag releases as `vX.Y.Z` to trigger packaging and artifact upload.

## Contact

Project maintainer: Shaik Baleeghuddin Kashif — kashif@sbkashif.com

# Speed

A simple CLI tool to measure internet speed (ping, download, and upload) using the
[ndt7 protocol](https://www.measurementlab.net/tests/ndt/) and [M-Lab servers](https://www.measurementlab.net/).

## Requirements

- Go 1.25 or later

## Usage

Clone the repository and build the binary:

```bash
git clone https://github.com/sibiraj-s/speed.git
cd speed
go run .
```

You will see output similar to:

```text
Retrieving speedtest.net configuration...

Server found: mlab1 at City, Country

↔ Ping (avg)    :      12 ms
↓ Download speed:   100.23 Mbps
↑ Upload speed  :    45.67 Mbps

🚀 Test complete!
```

## Install

You can also install the CLI directly using Go:

```bash
go install github.com/sibiraj-s/speed@latest
```

Or with [homebrew](https://brew.sh/)

```bash
brew tap sibiraj-s/speed https://github.com/sibiraj-s/speed
brew install --HEAD sibiraj-s/speed/speed
```

or with [mise](https://mise.jdx.dev/)

```bash
mise use go:github.com/sibiraj-s/speed
```

## How it works

- Uses [ndt7-client-go](https://github.com/m-lab/ndt7-client-go) to locate the nearest M-Lab server and perform speed
  tests
- Uses [pro-bing](https://github.com/prometheus-community/pro-bing) for ICMP ping

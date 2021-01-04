# Golang NBA Slack Bot

## Dependencies
`go version go1.13.8` or higher

## Installation
From the root of this project, first generate a private key

```
# Key considerations for algorithm "RSA" ≥ 2048-bit
openssl genrsa -out server.key 2048

```
Create a public self-signed public key from the private
```
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```
build the app and start
```
go build
SLACK_SS=YOUR_SIGNING_SECRET ./go_nba
```

## Deployment
On your distribution system of choice, configure your vms or containers to Pacific Stadard Time.
The application runs on port 8080.
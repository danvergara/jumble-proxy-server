# jumble-proxy-server

__This application is a proxy server used by the Jumble Nostr client as a workaround to fix CORS errors, so that the client can show the URL preview from given links Open Graph data.__

## Usage

### From source

Compile the binary:

```
CGO_ENABLED=0 go build -o bin/jumble-proxy-server .
```

Run the binary with environment variables:

```sh
ALLOW_ORIGIN=https://jumble.social PORT=8080 bin/jumble-proxy-server server
```

### Using Docker

Pull the image:

```
docker pull ghcr.io/danvergara/jumble-proxy-server:latest
```

Run the container;

```
docker run --rm -e PORT=8080 -p 8080:8080 ghcr.io/danvergara/jumble-proxy-server:latest
```

### Configuration

The configuration of the proxy server is done through environment variables. Most of them are optional.

- `PORT` Define the port the proxy server will be listening to (default: 8000)
- `ENABLE_PPROF` Enable pprof routes if present and equal to "true" (optional)
- `JUMBLE_PROXY_GITHUB_TOKEN` GitHub Token needed to authenticate with the GitHub API (mandatory) 

```
docker run --rm -e JUMBLE_PROXY_GITHUB_TOKEN=${JUMBLE_PROXY_GITHUB_TOKEN} -e PORT=8080 -p 8080:8080 ghcr.io/danvergara/jumble-proxy-server:latest
```

### How to hit the proxy server

The inner URL needs to be encoded so it doesn't break the outer URL structure.

```sh
curl -X GET http://localhost:8080/sites/https%3A%2F%2Fyoutu.be%2FNVm_jGdwTjQ%3Fsi%3DblYLT44WrrPjL9gU
```

The server will respond with the HTML from the website of interest.

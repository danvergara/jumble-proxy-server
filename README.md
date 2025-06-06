# jumble-proxy-server

__This application is a proxy server used by the Jumble Nostr client as a workaround to fix CORS erros,so that the client can show the URL preview from links' Open Graph data.__

## Usage
```sh
ALLOW_ORIGIN=https://jumble.social PORT=8080 jumble-proxy-server server
```

```sh
curl -X GET http://localhost:8080/sites/https%3A%2F%2Fyoutu.be%2FNVm_jGdwTjQ%3Fsi%3DblYLT44WrrPjL9gU
```

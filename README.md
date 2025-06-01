# jumble-proxy-server

__This application is a proxy server used by the Jumble Nostr client as a workaround to fix CORS erros,so that the client can show the URL preview from links' Open Graph data.__

## Usage
```sh
jumble-proxy-server -a https://jumble.social -p 8080
```

```sh
curl -X GET http://localhost:8080/sites/https:/www.youtube.com/watch\?v\=i-OZcpNSzg0
```

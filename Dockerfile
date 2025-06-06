FROM golang:1.24-bookworm AS build

ARG TARGETOS
ARG TARGETARCH

RUN apt-get update -y \
  && apt-get clean

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o jumble-proxy-server .

FROM gcr.io/distroless/base-debian11 AS build-release-stage

ENV PORT=8080
ENV AllOW_ORIGIN=https://jumble.social

COPY --from=build /app/jumble-proxy-server /bin/jumble-proxy-server

ENTRYPOINT ["/bin/jumble-proxy-server", "server"] 

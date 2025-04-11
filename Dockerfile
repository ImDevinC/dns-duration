FROM golang:1.24.1 AS build

WORKDIR /usr/src/app

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o dns-duration .

FROM gcr.io/distroless/static

COPY --from=build /usr/src/app/dns-duration /dns-duration

ENTRYPOINT ["/dns-duration"]

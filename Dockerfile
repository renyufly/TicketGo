FROM golang:1.27.0-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ticketgo ./cmd/server

FROM alpine:3.22
RUN addgroup -S ticketgo && adduser -S -G ticketgo ticketgo
USER ticketgo
COPY --from=build /out/ticketgo /usr/local/bin/ticketgo
EXPOSE 8080
ENTRYPOINT ["ticketgo"]


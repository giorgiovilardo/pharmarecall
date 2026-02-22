FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pharmarecall ./cmd/server

FROM scratch
COPY --from=build /pharmarecall /pharmarecall
EXPOSE 8080
ENTRYPOINT ["/pharmarecall"]

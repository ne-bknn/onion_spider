FROM golang:1.16-buster as build

WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY parser.go .

RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/parser

# Now copy it into our base image.
FROM gcr.io/distroless/base-debian10

COPY --from=build /go/bin/parser /

USER nobody
CMD ["/parser"]

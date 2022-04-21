FROM golang:1.16-alpine

WORKDIR /app/chat-telnet
RUN mkdir -p /app/log/

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

EXPOSE 9000

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o chat-telnet .

CMD ["sh", "./init_app"]

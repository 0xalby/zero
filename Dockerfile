FROM golang:1.23.3

WORKDIR /app

COPY . /app

RUN go mod tidy
RUN go build -o bin/zero -ldflags="-s -w" .

EXPOSE 6000

CMD ["./bin/zero"]

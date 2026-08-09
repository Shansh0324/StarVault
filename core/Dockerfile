FROM golang:alpine

WORKDIR /app

COPY go.mod ./
COPY . .
RUN go get github.com/lib/pq && go get golang.org/x/crypto/bcrypt && go mod tidy

RUN go build -o core-service main.go

EXPOSE 8080

CMD ["./core-service"]

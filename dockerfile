FROM golang:alpine
WORKDIR /app
COPY . /app
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o main.out -mod=vendor main.go
CMD [ "./main.out" ]

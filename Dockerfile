FROM golang:1.24-alpine

WORKDIR /app
COPY . .

RUN go build -o main ./cmd/app
RUN chmod +x main

EXPOSE 8080
CMD ["./main"]

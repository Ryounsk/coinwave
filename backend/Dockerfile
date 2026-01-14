FROM mirror.gcr.io/library/golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main main.go

# Final stage
FROM mirror.gcr.io/library/alpine:latest

WORKDIR /app

# Install sqlite libs if needed (though we use mysql now, keep for compatibility or remove)
RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]

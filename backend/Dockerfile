FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

# Force rebuild to pick up code changes
ENV REBUILD_DATE="2026-01-19-04"

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main main.go

# Final stage
FROM golang:1.23-alpine AS final

WORKDIR /app

# Install sqlite libs if needed (though we use mysql now, keep for compatibility or remove)
RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]

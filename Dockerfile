FROM golang:1.25-alpine AS builder
WORKDIR /app
# 使用中国可访问的 Go 模块代理（proxy.golang.org 被 GFW 封锁）
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /server .
COPY migrations/ ./migrations/
EXPOSE 16080
CMD ["./server"]
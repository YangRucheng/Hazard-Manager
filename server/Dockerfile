# 后端镜像：多阶段构建，Go 静态编译 + Alpine 运行时
FROM golang:1.27-alpine AS builder

WORKDIR /build

# 先拷贝依赖清单，利用层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hazard-server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    mkdir -p /app/data/uploads && \
    adduser -D -u 10001 appuser

WORKDIR /app
COPY --from=builder /out/hazard-server /app/hazard-server

# 运行时数据（图片上传）目录
VOLUME ["/app/data"]
ENV UPLOAD_DIR=/app/data/uploads
ENV SERVER_PORT=8090

USER appuser
EXPOSE 8090

ENTRYPOINT ["/app/hazard-server"]
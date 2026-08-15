# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/taqlyn-mcp ./cmd/taqlyn-mcp

FROM alpine:3.21
RUN adduser -D -u 65532 mcp \
  && apk add --no-cache ca-certificates wget
USER mcp
COPY --from=build /out/taqlyn-mcp /usr/local/bin/taqlyn-mcp
ENV TAQLYN_MCP_TRANSPORT=http \
    TAQLYN_MCP_ADDR=:8787 \
    TAQLYN_API_URL=http://127.0.0.1:8080
EXPOSE 8787
HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8787/healthz >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/taqlyn-mcp"]

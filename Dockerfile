# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/taqlyn-mcp ./cmd/taqlyn-mcp

FROM alpine:3.21
RUN adduser -D -u 65532 mcp \
  && apk add --no-cache ca-certificates wget \
  && mkdir -p /workspace /home/mcp/.config/taqlyn \
  && chown -R mcp:mcp /workspace /home/mcp
USER mcp
WORKDIR /workspace
COPY --from=build /out/taqlyn-mcp /usr/local/bin/taqlyn-mcp
ENV TAQLYN_MCP_TRANSPORT=stdio \
    TAQLYN_MCP_ADDR=:8787 \
    TAQLYN_WORKSPACE=/workspace \
    TAQLYN_API_URL=https://api.rutvik.qzz.io
EXPOSE 8787
ENTRYPOINT ["/usr/local/bin/taqlyn-mcp"]

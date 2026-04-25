FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/mcp-server ./cmd/mcp-server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/mcp-server /app/mcp-server

HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD ["/app/mcp-server", "--healthcheck"]

ENTRYPOINT ["/app/mcp-server"]

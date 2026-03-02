# Stage 1: Frontend build
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Backend build
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /clabnoc ./cmd/clabnoc

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tcpdump
COPY --from=backend-builder /clabnoc /usr/local/bin/clabnoc

EXPOSE 8080
ENTRYPOINT ["clabnoc"]

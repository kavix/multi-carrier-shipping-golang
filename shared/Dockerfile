FROM golang:1.21-alpine AS shared-base
WORKDIR /shared
COPY go.mod go.sum ./
RUN go mod download
COPY . .

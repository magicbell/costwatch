# Base image for Go services (api, worker, seed)
FROM golang:1.24
WORKDIR /app

# Pre-download Go modules for faster cold starts
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the repo
COPY . .

# No default CMD; docker-compose will provide commands per service

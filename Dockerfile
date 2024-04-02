FROM golang:1.22.0 AS builder

WORKDIR /go/src/yle-bot

# Copy only the necessary files and folders excluding those listed in .dockerignore
COPY . .

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /yle-bot

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy only the necessary files and the db folder from the builder stage
COPY --from=builder /yle-bot .

CMD ["./yle-bot"]
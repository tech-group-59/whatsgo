FROM golang:1.22.0
LABEL version="1.0"
LABEL description="What's Go application"
LABEL maintainer="Dmytro Karpovych <karpovych.d.v@gmail.com>"

RUN apt-get update -qq

# Setup project
WORKDIR /app

# Create a user and switch to it
RUN adduser --disabled-password --gecos '' appuser
USER appuser

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Change ownership of the build directory to appuser
USER root
RUN mkdir ./build
RUN chown -R appuser:appuser ./build
USER appuser

# Build the application
RUN go build -o build/whatsgo ./cmd/whatsgo

CMD ["./build/whatsgo", "--config=config/config.yaml", "-detached"]

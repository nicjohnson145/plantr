FROM golang:1.23-alpine AS builder

# Install task
RUN mkdir /tmp/task && \
    wget https://github.com/go-task/task/releases/download/v3.20.0/task_linux_amd64.tar.gz && \
    tar -xzf task_linux_amd64.tar.gz --directory /tmp/task && \
    mv /tmp/task/task /bin/task && \
    rm -rf /tmp/task task_linux_amd64.tar.gz

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN task build-controller

FROM alpine:latest
COPY --from=builder /src/plantr-controller /bin/plantr-controller
ENTRYPOINT ["/bin/plantr-controller"]

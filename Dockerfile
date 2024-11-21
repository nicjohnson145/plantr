ARG BINARY_NAME
ARG TASK

FROM golang:1.23-alpine AS builder

# Install task
RUN mkdir /tmp/task && \
    wget https://github.com/go-task/task/releases/download/v3.20.0/task_linux_amd64.tar.gz && \
    tar -xzf task_linux_amd64.tar.gz --directory /tmp/task && \
    mv /tmp/task/task /bin/task && \
    rm -rf /tmp/task task_linux_amd64.tar.gz

RUN apk add git ca-certificates

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ARG TASK
ARG BINARY_NAME

RUN task $TASK

FROM scratch
ARG BINARY_NAME
COPY --from=builder /src/${BINARY_NAME} /bin/entrypoint
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/entrypoint"]

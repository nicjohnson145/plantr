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

RUN task build-agent

FROM ubuntu:24.04
RUN apt update && apt install -y ca-certificates sudo git curl build-essential
RUN useradd -ms /bin/bash newuser
RUN groupadd passwordless
RUN usermod -a -G passwordless newuser
RUN echo "newuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
WORKDIR /home/newuser
COPY hack/functional-test/test.sh test.sh
RUN chmod +x /home/newuser/test.sh
USER newuser
COPY --from=builder /src/plantr-agent /bin/plantr-agent
ENV PATH="/usr/local/go/bin:$PATH"
# Install homebrew (duurrr script from the internet but :shrug:)
ENV PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:$PATH"
RUN NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
ENTRYPOINT ["/bin/plantr-agent"]

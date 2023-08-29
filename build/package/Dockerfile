FROM fedora:38

ARG WIREGUARD_CONFIG_FILENAME

# Install tools
# RUN dnf update -y
RUN dnf install wireguard-tools -y

# Install go
ENV GO_FILENAME="go1.21.0.linux-amd64.tar.gz"
WORKDIR /tmp
RUN curl -LO https://go.dev/dl/${GO_FILENAME}
RUN tar -C /usr/local -xzvf ${GO_FILENAME}
ENV PATH=$PATH:/usr/local/go/bin
RUN go version

# Setup wireguard
RUN wg-quick up ${WIREGUARD_CONFIG_FILENAME}
RUN wg

# Setup
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod verify
RUN go mod tidy

# Copy source
COPY data .
COPY Makefile .
COPY main.go .
COPY internal .

# Build and run
RUN make build
RUN make run
# RUN echo "hej" | less

# RUN go build -o $(make out)
# CMD [ "./$(make out)" ]

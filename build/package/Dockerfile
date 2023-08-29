FROM fedora:38

ARG PROTONVPN_USERNAME
ARG PROTONVPN_PASSWORD

# Update

# Setup the Proton VPN repository
WORKDIR /tmp
ENV PROTONVPN_FILENAME="protonvpn-stable-release-1.0.1-2.noarch.rpm"
RUN curl -LO https://repo.protonvpn.com/fedora-38-stable/protonvpn-stable-release/${PROTONVPN_FILENAME}
RUN dnf install ./${PROTONVPN_FILENAME}
RUN dnf update -y
RUN dnf install protonvpn-cli -y

# Login
RUN cat ${PROTONVPN_PASSWORD} | protonvpn-cli login ${PROTONVPN_USERNAME}

# Connect to Proton VPN servers
RUN protonvpn-cli connect --protocol udp --random

# Install go
ENV GO_FILENAME="go1.21.0.linux-amd64.tar.gz"
WORKDIR /tmp
RUN curl -LO https://go.dev/dl/${GO_FILENAME}
RUN tar -C /usr/local -xzvf ${GO_FILENAME}
ENV PATH=$PATH:/usr/local/go/bin
RUN go version

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

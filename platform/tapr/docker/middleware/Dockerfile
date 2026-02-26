
# Build the manager binary
FROM golang:1.24.0 as builder
WORKDIR /workspace
RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    apt-get update && \
    wget https://github.com/mongodb-js/mongosh/releases/download/v1.6.2/mongodb-mongosh_1.6.2_${arch}.deb && \
    apt-get install ./mongodb-mongosh_1.6.2_${arch}.deb

RUN apt-get update && apt-get install -y libkrb5-3 \
        libssl-dev \
        libcurl4 \
        libgssapi-krb5-2 \
        libuuid1 \
        libc6 \
        libgcc1 \
        libcom-err2 \
        libcomerr2 \
        keyutils && \
    rm -rf /var/lib/apt/lists/*


COPY go.mod go.sum ./
RUN \
	echo ">> Downloading go modules..." && \
	go mod download

COPY . .
RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    if [ $arch = "arm64" ]; then \
        LIBP=aarch64; \
    else \
        LIBP=x86_64; \
    fi && \
    echo "Library path is set to $LIBP" && \
    cp /usr/lib/${LIBP}-linux-gnu/libkrb5support.so.0 /tmp && \
    cp /usr/lib/${LIBP}-linux-gnu/libk5crypto.so.3 /tmp && \
    cp /usr/lib/${LIBP}-linux-gnu/libgssapi_krb5.so.2 /tmp && \
    cp /usr/lib/${LIBP}-linux-gnu/libkrb5.so.3 /tmp && \
    cp /lib/${LIBP}-linux-gnu/libcom_err.so.2 /tmp && \
    cp /lib/${LIBP}-linux-gnu/libgcc_s.so.1 /tmp && \
    cp /lib/${LIBP}-linux-gnu/libkeyutils.so.1 /tmp


RUN go build -ldflags="-s -w" -o middleware-operator cmd/middleware/main.go


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/base:nonroot
FROM gcr.io/distroless/base:debug
WORKDIR /

COPY --from=builder /bin/sh /bin/sh
COPY --from=builder /usr/bin/arch /bin/arch
COPY --from=builder /tmp /tmp

RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    if [ $arch = "arm64" ]; then \
       LIBP=aarch64; \
    else \
       LIBP=x86_64; \
    fi && \
    echo "Library path is set to $LIBP" && \
    cp /tmp/* /usr/lib/${LIBP}-linux-gnu && \
    rm -rf /tmp/* /bin/sh /bin/arch

COPY --from=builder /workspace/middleware-operator .
COPY --from=builder /usr/bin/mongosh /usr/bin

EXPOSE 8080

ENTRYPOINT ["/middleware-operator"]

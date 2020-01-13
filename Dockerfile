# Stage 1 - build the executable.
FROM golang:1-buster AS stage1

WORKDIR /build/src
COPY ./ ./
ENV GOBIN="/build/bin"
RUN ["go", "install", "-ldflags=-extldflags \"-static\"", "go.spiff.io/dummysv"]

# Stage 2 - final docker image.
FROM debian:buster AS stage2

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/dummysv", "-L=:8080"]

# Copy executable.
COPY --from=stage1 /build/bin/dummysv /usr/local/bin/

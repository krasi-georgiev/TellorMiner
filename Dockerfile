FROM golang:buster AS builder
WORKDIR /go/src
RUN apt-get update
RUN apt-get install -y --no-install-recommends ocl-icd-opencl-dev
COPY ./ .
RUN make build GIT_TAG=. GIT_HASH=.

# Can't use alpine because of the ocl-icd-opencl-dev requirement fr GPU mining.
FROM debian:buster-slim  
WORKDIR /root/
RUN apt-get update
RUN apt-get install -y --no-install-recommends ocl-icd-opencl-dev && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/src/tellor .
CMD ["./tellor"]  
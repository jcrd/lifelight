FROM golang:1-bullseye

LABEL os=linux
LABEL arch=arm

ENV GOOS=linux
ENV GOARCH=arm
ENV CGO_ENABLED=1
ENV CC=arm-linux-gnueabi-gcc-10
ENV CXX=arm-linux-gnueabi-g++-10

RUN dpkg --add-architecture arm
RUN apt update
RUN apt install -y --no-install-recommends \
        libstdc++6-armel-cross \
        g++-10-arm-linux-gnueabi

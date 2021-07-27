FROM golang:buster

WORKDIR /workdir

RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*

COPY . /workdir/

RUN go test

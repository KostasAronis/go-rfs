FROM golang:1.15.6 AS Builder

RUN mkdir /builddir

ADD . /build

WORKDIR /build

RUN ls /build

RUN go build -o /output/miner cmd/miner/miner.go

FROM alpine:latest

COPY --from=Builder /output/miner .

CMD ["./miner"]


# TO BUILD:
# docker build -f miner.dockerfile --tag miner .
# TO RUN:
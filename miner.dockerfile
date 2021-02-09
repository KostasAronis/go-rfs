FROM golang:1.15.6 AS Builder

RUN mkdir /builddir

ADD . /build

WORKDIR /build

RUN ls /build

RUN CGO_ENABLED=0 GOOS=linux go build -o /output/miner -a -ldflags '-extldflags "-static"' cmd/miner/miner.go

FROM alpine:latest

EXPOSE 8001
EXPOSE 9001

COPY --from=Builder /output/miner .

CMD ["./miner"]


# TO BUILD:
# docker build -f miner.dockerfile --tag miner .
# TO RUN:
FROM golang:alpine as build
ADD  . /go/src/github.com/tjamet/sqs-gc
RUN go build -o /bin/sqs-gc github.com/tjamet/sqs-gc

FROM alpine
COPY --from=build /bin/sqs-gc /bin/sqs-gc
ENTRYPOINT ["/bin/sqs-gc"]
CMD ["--help"]

FROM alpine:20200428 AS development

ENV PROJECT_PATH=/network-qos
ENV PATH=$PATH:$PROJECT_PATH/build
ENV CGO_ENABLED=0
ENV GO_EXTRA_BUILD_ARGS="-a -installsuffix cgo"

RUN apk add --no-cache go ca-certificates make git bash alpine-sdk

RUN mkdir -p $PROJECT_PATH
COPY . $PROJECT_PATH
WORKDIR $PROJECT_PATH

RUN make build

FROM alpine:20200428 AS production

EXPOSE 8080
WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=development /network-qos/build/network-qos-service .
ENTRYPOINT ["./network-qos-service"]

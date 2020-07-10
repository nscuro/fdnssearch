FROM golang:1.14-alpine as build
RUN apk --no-cache add git
WORKDIR /opt/fdnssearch
COPY . .
RUN GO111MODULE=on go build -v ./cmd/fdnssearch

FROM alpine:latest
COPY --from=build /opt/fdnssearch/fdnssearch /bin/fdnssearch
ENTRYPOINT [ "/bin/fdnssearch" ]

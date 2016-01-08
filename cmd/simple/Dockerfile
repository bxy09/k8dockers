FROM daocloud.io/gpx_dev/go:1.5.1
ADD . /go/src/cmd/
RUN cd /go/src/cmd/;go get ./...;go install
CMD /go/bin/server -f /go/src/cmd/sample.yaml

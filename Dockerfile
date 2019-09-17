FROM golang:1.9

ADD main.go /root

WORKDIR /root

RUN go build -v -o hostpathperpod

FROM golang:1.9-alpine

ADD deploy.sh /usr/bin

RUN chmod +x /usr/bin/deploy.sh

COPY --from=0 /root/hostpathperpod /hostpathperpod

CMD ["/usr/bin/deploy.sh"]

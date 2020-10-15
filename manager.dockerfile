FROM golang:latest
USER 0
WORKDIR /opt
ENV GOPATH ${GOPATH}:/opt
ADD * /opt/
RUN go mod download
RUN go build -o manager
RUN chmod +x manager
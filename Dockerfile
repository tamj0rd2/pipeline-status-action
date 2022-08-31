# Container image that runs your code
FROM golang:1.19-alpine

WORKDIR /our-code
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY main.go ./main.go
COPY github ./github
COPY slack ./slack

RUN go build -o ./github-action main.go

# Code file to execute when the docker container starts up (`entrypoint.sh`)
ENTRYPOINT ["/our-code/github-action"]
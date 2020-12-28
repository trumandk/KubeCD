FROM golang:alpine AS builder

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache curl
WORKDIR /
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
WORKDIR /app/

RUN go get github.com/go-git/go-git
RUN go get github.com/sparrc/go-ping
RUN go get github.com/docker/docker/client
RUN go get github.com/tidwall/sjson
RUN go get github.com/shirou/gopsutil

COPY main.go main.go
RUN CGO_ENABLED=0 go build -o /main

FROM alpine
WORKDIR /
COPY --from=builder /kubectl /kubectl
RUN chmod +x /kubectl
ENV SSH_KNOWN_HOSTS=/root/.ssh/known_hosts
COPY --from=builder /main /main
ENTRYPOINT ["/main"]
#ENTRYPOINT ["/kubectl", "get", "pods"]

# stage 1
FROM golang:1.11.2 AS stage-one-env
COPY ./ /
WORKDIR /go
# build first stage
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mainapp ./src

# stage 2
FROM alpine
COPY --from=stage-one-env /go/src/github/bbtran/go-rest-api/mainapp ./

EXPOSE 8080 

ENTRYPOINT ["./mainapp"]

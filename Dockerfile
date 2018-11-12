# stage 1
FROM golang:1.11.2 AS stage-one-env
# add dep depenpendency manager
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep
COPY Gopkg.toml Gopkg.lock ./

COPY . ./

# check dependencies
RUN dep ensure --vendor-only

# build first stage
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mainapp ./src

# stage 2
FROM alpine
COPY --from=stage-one-env /mainapp ./

EXPOSE 8080 

ENTRYPOINT ["./mainapp"]

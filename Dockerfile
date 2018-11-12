# stage 1
FROM golang:1.11.2 AS stage-one-env
COPY . ./
# build first stage
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mainapp ./src

# stage 2
FROM alpine
COPY --from=stage-one-env /mainapp ./

EXPOSE 8080 

ENTRYPOINT ["./mainapp"]

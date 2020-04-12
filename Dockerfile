FROM golang:1.13.9 as buildEnv
WORKDIR /app
COPY . .
# RUN apt-get update && apt-get install tree 
# RUN tree
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-c -w -s -linkmode external -extldflags -static"


FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
ENV PATH=/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app/_files
COPY --from=buildEnv /app/_files/. /app/_files/.
WORKDIR /app/_templates
COPY --from=buildEnv /app/_templates/. /app/_templates/.
WORKDIR /app
COPY --from=buildEnv /app/. /app/.
###### main is the module name, update it if another module name is specified in go.mod
ENTRYPOINT ["/app/main"]

# FROM golang:1.13.9
# WORKDIR /app
# COPY --from=0 /app/. /app/.
# RUN apt-get update && apt-get install tree 
# RUN tree
# ENTRYPOINT ["/app/main"]


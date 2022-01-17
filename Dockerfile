FROM golang:1.17
WORKDIR /build
COPY *.go ./
COPY go.* ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o grouper .
RUN strip grouper

FROM alpine:latest
RUN apk update
WORKDIR /app
COPY --from=0 /build/grouper ./
EXPOSE 8080
CMD ["./grouper"]

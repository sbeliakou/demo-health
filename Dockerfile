#
# docker build -t sbeliakou/playpit-liveness-demo .
#

FROM golang:1.22.1 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM scratch
COPY --from=builder /app/main /
COPY --from=builder /app/index.html /
ENV PORT 8080
EXPOSE 8080

CMD ["/main"]

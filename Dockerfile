FROM golang:alpine as builder 
RUN mkdir /build
Add . /build/
WORKDIR /build
RUN go build .

FROM alpine
COPY --from=builder /build/stratum-health /app/
WORKDIR /app
CMD ["./stratum-health"]
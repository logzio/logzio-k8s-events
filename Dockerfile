FROM golang:1.19-alpine AS build

LABEL authors="ralongit"
WORKDIR /app

COPY . .

RUN go build -o main .

FROM alpine:3.14
COPY --from=build /app/main /app/main
CMD ["/app/main"]
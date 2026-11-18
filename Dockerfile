FROM golang:1.27.4-alpine3.24 AS build

LABEL authors="ralongit"
WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o main .

FROM alpine:3.24

# A numeric UID lets Kubernetes' runAsNonRoot verify the user isn't root
RUN addgroup -S -g 10001 appgroup && adduser -S -u 10001 -G appgroup appuser

COPY --from=build /app/main /app/main

USER 10001
CMD ["/app/main"]

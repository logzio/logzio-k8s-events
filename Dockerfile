FROM golang:1.27.4-alpine3.24 AS build

LABEL authors="ralongit"
WORKDIR /app

COPY . .

RUN go build -v -o main .

# Create a non root group and user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Switch to the non root new user
USER appuser

FROM alpine:3.24
COPY --from=build /app/main /app/main
CMD ["/app/main"]
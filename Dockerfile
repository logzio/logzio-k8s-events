FROM golang:1.19-alpine AS build

LABEL authors="ralongit"
WORKDIR /app

COPY . .

# Create a non root group and user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Switch to the non root new user
USER appuser

RUN go build -o main .

FROM alpine:3.14
COPY --from=build /app/main /app/main
CMD ["/app/main"]
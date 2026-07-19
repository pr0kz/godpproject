FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /service ./services/shop/cmd
FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /service /service
EXPOSE 8080
ENTRYPOINT ["/service"]


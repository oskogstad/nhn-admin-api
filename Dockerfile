FROM golang:1.15.6-alpine AS build
WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' -a \
    -o /app/admin-api .

FROM alpine:3.12.1 AS bin
COPY --from=build /app/admin-api /app/admin-api
CMD [ "/app/admin-api" ]
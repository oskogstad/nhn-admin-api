FROM golang:1.15.6-alpine AS build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' -a \
    -o /go/admin-api .

FROM alpine:3.12.1 AS bin
COPY --from=build /go/admin-api /go/admin-api
CMD [ "/go/admin-api" ]
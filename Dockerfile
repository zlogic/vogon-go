FROM golang:1.12-alpine as builder

# Create app directory
RUN mkdir -p /usr/src/vogon
WORKDIR /usr/src/vogon

# Bundle app source
COPY . /usr/src/vogon

# Install build dependencies
RUN apk add --no-cache --update build-base git

# Run tests
RUN go test ./...

# Build app
RUN go build -ldflags="-s -w" && \
  mkdir /usr/src/vogon/dist && \
  cp -r vogon-go static templates /usr/src/vogon/dist

# Copy into a fresh image
FROM alpine:3.9

COPY --from=builder /usr/src/vogon/dist /usr/local/vogon

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /usr/local/vogon

EXPOSE 8080
CMD [ "./vogon-go" ]
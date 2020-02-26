FROM golang:1.14-alpine as builder

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
RUN CGO_ENABLED=0 go build -ldflags="-s -w" && \
  mkdir /usr/src/vogon/dist && \
  cp -r vogon-go static templates /usr/src/vogon/dist

# Create non-root user
RUN adduser -D -H -u 10001 vogon

# Copy into a fresh image
FROM scratch

COPY --from=builder /usr/src/vogon/dist /usr/local/vogon
COPY --from=builder /etc/passwd /etc/passwd

WORKDIR /usr/local/vogon

EXPOSE 8080
USER vogon
CMD [ "./vogon-go" ]

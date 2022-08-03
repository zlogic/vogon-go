FROM golang:1.19-alpine as builder

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
  cp vogon-go /usr/src/vogon/dist

# Create tmp for multipart uploads
RUN mkdir -p /usr/src/vogon/tmp && \
  chmod ug+rwx /usr/src/vogon/tmp

# Copy into a fresh image
FROM scratch

COPY --from=builder /usr/src/vogon/dist /usr/local/vogon
COPY --from=builder --chown=1001:1001 /usr/src/vogon/tmp /

WORKDIR /usr/local/vogon

EXPOSE 8080
USER 1001
CMD [ "./vogon-go" ]

FROM alpine:latest as builder

# Create non-root user
RUN adduser -D -H -u 10001 vogon

# Copy build artifact
COPY . /usr/src/vogon

# Create bin directory
RUN mkdir -p /usr/local/vogon && \
  cd /usr/src/vogon && \
  cp -r vogon-go static templates /usr/local/vogon

# Copy into a fresh image
FROM scratch

COPY --from=builder /usr/local/vogon /usr/local/vogon
COPY --from=builder /etc/passwd /etc/passwd

WORKDIR /usr/local/vogon

EXPOSE 8080
USER vogon
CMD [ "./vogon-go" ]

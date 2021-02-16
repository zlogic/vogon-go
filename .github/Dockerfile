FROM alpine:latest as builder

# Copy build artifact
COPY . /usr/src/vogon

# Create bin directory
RUN mkdir -p /usr/local/vogon && \
  cd /usr/src/vogon && \
  cp -r vogon-go /usr/local/vogon

# Copy into a fresh image
FROM scratch

COPY --from=builder /usr/local/vogon /usr/local/vogon

WORKDIR /usr/local/vogon

EXPOSE 8080
USER 1001
CMD [ "./vogon-go" ]

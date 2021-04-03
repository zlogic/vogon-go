FROM alpine:latest as builder

# Copy build artifact
COPY . /usr/src/vogon

# Create bin directory
RUN mkdir -p /usr/local/vogon && \
  cd /usr/src/vogon && \
  cp -r vogon-go /usr/local/vogon

# Create tmp for multipart uploads
RUN mkdir -p /usr/src/vogon/tmp && \
  chmod ug+rwx /usr/src/vogon/tmp

# Copy into a fresh image
FROM scratch

COPY --from=builder /usr/local/vogon /usr/local/vogon
COPY --from=builder --chown=1001:1001 /usr/src/vogon/tmp /

WORKDIR /usr/local/vogon

EXPOSE 8080
USER 1001
CMD [ "./vogon-go" ]

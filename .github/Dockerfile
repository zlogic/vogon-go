FROM scratch

COPY --chown=1001:1001 vogon-go /usr/local/vogon/
COPY tmp /tmp

WORKDIR /usr/local/vogon

EXPOSE 8080
USER 1001
CMD [ "./vogon-go" ]

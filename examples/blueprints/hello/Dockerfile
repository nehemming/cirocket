FROM alpine:3.13
WORKDIR /opt/app
COPY oauthproxy .
USER 1000:1000
# executable
ENTRYPOINT [ "./{{.project_name}}" ]
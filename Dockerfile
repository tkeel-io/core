#
# Qingcloud IotHub Metadata Dockerfile
#
FROM alpine:3.13
RUN mkdir /keel
ADD dist/linux_amd64/release/core /keel
ADD config.yml /keel
WORKDIR /keel
CMD ["/keel/core", "serve", "--config", "config.yml"]


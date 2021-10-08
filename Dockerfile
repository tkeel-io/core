#
# Qingcloud IotHub Metadata Dockerfile
#
FROM alpine:3.13
RUN mkdir /keel
ADD bin/linux/core /keel
ADD config.yml /keel
WORKDIR /keel
CMD ["/keel/core", "serve", "--config", "config.yml"]


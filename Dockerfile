FROM alpine:3.4
RUN apk add --no-cache ca-certificates
RUN wget -O- "http://s3.amazonaws.com/babl/babl-server_linux_amd64.gz" | gunzip > /bin/babl-server && chmod +x /bin/babl-server
ADD events_linux_amd64 /bin/app
ADD subscriptions.json /
ADD start /bin/
CMD ["start"]

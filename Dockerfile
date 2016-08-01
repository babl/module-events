FROM busybox
RUN wget -O- "http://s3.amazonaws.com/babl/babl-server_linux_amd64.gz" | gunzip > /bin/babl-server && chmod +x /bin/babl-server
ADD events_linux_amd64 /bin/app
ADD subscriptions.json /
CMD ["babl-server"]

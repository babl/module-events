FROM busybox
RUN wget -O- "http://s3.amazonaws.com/babl/babl-server_linux_amd64.gz" | gunzip > /bin/babl-server && chmod +x /bin/babl-server
ADD app /bin/app
RUN chmod +x /bin/app
CMD ["babl-server"]

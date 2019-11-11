FROM alpine:latest
ENV WORKDIR /workdir

WORKDIR $WORKDIR
RUN mkdir -p $WORKDIR/etc/
COPY syslogmonitor ./

CMD ["sh", "-c", "./syslogmonitor -f=$WORKDIR/etc/config.yml"]
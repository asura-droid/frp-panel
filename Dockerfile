FROM alpine

ARG ARCH

RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.18/main" > /etc/apk/repositories && \
    apk update && \
    apk add --no-cache curl bash sqlite

ENV TZ Asia/Shanghai

WORKDIR /app
COPY ./frp-panel-${ARCH} /app/frp-panel
COPY ./etc /app/etc

RUN ln -sf /app/etc/Shanghai /etc/localtime && \
    mv /app/etc/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt && \
    mkdir -p /data

EXPOSE 9000
EXPOSE 9001

ENV DB_DSN /data/data.db

ENTRYPOINT [ "/app/frp-panel" ]
CMD [ "master" ]

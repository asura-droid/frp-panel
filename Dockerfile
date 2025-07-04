FROM debian:bullseye-slim

ARG ARCH

RUN apt update && \
    apt install -y curl bash sqlite3 ca-certificates && \
    apt clean

ENV TZ Asia/Shanghai

WORKDIR /app
COPY ./frp-panel-${ARCH} /app/frp-panel
COPY ./etc /app/etc

RUN ln -sf /app/etc/Shanghai /etc/localtime && \
    cp /app/etc/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt && \
    mkdir -p /data

EXPOSE 9000
EXPOSE 9001

ENV DB_DSN /data/data.db

ENTRYPOINT [ "/app/frp-panel" ]
CMD [ "master" ]

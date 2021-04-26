FROM alpine
RUN apk update && apk add -u --no-cache zip tini bash curl

COPY entrypoint.sh /usr/bin/
RUN chmod +x /usr/bin/entrypoint.sh

ADD bin/harvester-sb-collector.sh /usr/bin
RUN chmod +x /usr/bin/harvester-sb-collector.sh

ADD bin/support-bundle-utils /usr/bin
RUN chmod +x /usr/bin/support-bundle-utils

ENTRYPOINT ["entrypoint.sh"]

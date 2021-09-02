FROM alpine
RUN apk update && apk add -u --no-cache zip tini bash curl

COPY package/entrypoint.sh /usr/bin/
RUN chmod +x /usr/bin/entrypoint.sh

ADD hack/support-bundle-collector.sh /usr/bin
RUN chmod +x /usr/bin/support-bundle-collector.sh

ADD bin/support-bundle-kit /usr/bin
RUN chmod +x /usr/bin/support-bundle-kit

ADD hack/collector-* /usr/bin/
RUN chmod +x /usr/bin/collector-*

ENTRYPOINT ["entrypoint.sh"]

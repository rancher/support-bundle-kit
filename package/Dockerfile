# syntax=docker/dockerfile:1.7.0

FROM registry.suse.com/bci/bci-base:15.5

ARG TARGETPLATFORM

RUN if [ "$TARGETPLATFORM" != "linux/amd64" ] && [ "$TARGETPLATFORM" != "linux/arm64" ]; then \
    echo "Error: Unsupported TARGETPLATFORM: $TARGETPLATFORM" && \
    exit 1; \
    fi

RUN zypper -n rm container-suseconnect && \
    zypper -n install zip curl which tar && \
    zypper -n clean -a && rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/*

ENV ARCH=${TARGETPLATFORM#linux/}
ENV TINI_VERSION v0.19.0
ENV TINI_URL_amd64=https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini \
    TINI_URL_arm64=https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-arm64 \
    TINI_URL_s390x=https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-s390x \
    TINI_URL=TINI_URL_${ARCH}

RUN curl -sLf ${!TINI_URL} > /usr/bin/tini && chmod +x /usr/bin/tini && \
    curl -sfL https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${ARCH} -o /usr/bin/yq && chmod +x /usr/bin/yq

COPY package/entrypoint.sh /usr/bin/
RUN chmod +x /usr/bin/entrypoint.sh

ADD hack/support-bundle-collector.sh /usr/bin
RUN chmod +x /usr/bin/support-bundle-collector.sh

ADD bin/support-bundle-kit-${ARCH} /usr/bin/support-bundle-kit
RUN chmod +x /usr/bin/support-bundle-kit

ADD hack/collector-* /usr/bin/
RUN chmod +x /usr/bin/collector-*

ADD hack/common /tmp/

ENTRYPOINT ["entrypoint.sh"]

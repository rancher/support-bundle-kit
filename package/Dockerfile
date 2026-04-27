# syntax=docker/dockerfile:1.7.0

FROM registry.suse.com/bci/bci-base:15.7@sha256:3267acc7b9744218250e91e1d2d3176a113d88c4c55158945beb46150c0c7df2

ARG TARGETPLATFORM

RUN if [ "$TARGETPLATFORM" != "linux/amd64" ] && [ "$TARGETPLATFORM" != "linux/arm64" ]; then \
    echo "Error: Unsupported TARGETPLATFORM: $TARGETPLATFORM" && \
    exit 1; \
    fi

RUN zypper -n rm container-suseconnect && \
    zypper -n install zip curl which tar procps && \
    zypper -n clean -a && rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/*

ENV ARCH=${TARGETPLATFORM#linux/}

# renovate: datasource=github-releases depName=krallin/tini
ENV TINI_VERSION=v0.19.0
ENV TINI_SUM_amd64=93dcc18adc78c65a028a84799ecf8ad40c936fdfc5f2a57b1acda5a8117fa82c
ENV TINI_SUM_arm64=07952557df20bfd2a95f9bef198b445e006171969499a1d361bd9e6f8e5e0e81
ENV TINI_URL_amd64=https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini \
    TINI_URL_arm64=https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-arm64 \
    TINI_URL=TINI_URL_${ARCH}

# renovate: datasource=github-releases depName=mikefarah/yq
ENV YQ_VERSION=v4.53.2
ENV YQ_SUM_amd64=d56bf5c6819e8e696340c312bd70f849dc1678a7cda9c2ad63eebd906371d56b
ENV YQ_SUM_arm64=03061b2a50c7a498de2bbb92d7cb078ce433011f085a4994117c2726be4106ea

RUN eval "TINI_EXPECTED=\${TINI_SUM_${ARCH}}" && \
    curl -sLf "${!TINI_URL}" -o /usr/bin/tini && \
    echo "${TINI_EXPECTED}  /usr/bin/tini" | sha256sum -c - && \
    chmod +x /usr/bin/tini && \
    eval "YQ_EXPECTED=\${YQ_SUM_${ARCH}}" && \
    curl -sfL "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_${ARCH}" \
        -o /usr/bin/yq && \
    echo "${YQ_EXPECTED}  /usr/bin/yq" | sha256sum -c - && \
    chmod +x /usr/bin/yq

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

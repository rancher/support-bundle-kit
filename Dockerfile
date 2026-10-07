FROM registry.suse.com/bci/golang:1.26 AS builder

ARG MK_HOST_ARCH
ENV ARCH=$MK_HOST_ARCH

ENV GOLANGCI_LINT_VERSION=v2.11.4

# -- for make rules
## install docker client
RUN zypper -n install ca-certificates awk lsb-release rsync docker containerd

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" $GOLANGCI_LINT_VERSION

# ---- base ----
FROM builder AS base
WORKDIR /go/src/github.com/rancher/support-bundle-kit

# to exclude some files, add them in .dockerignore
COPY . .

# ---- build ----
FROM base AS build
ARG MK_REPO_ID

RUN --mount=type=cache,target=/go/pkg/mod,id=support-bundle-kit-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/github.com/rancher/support-bundle-kit/.cache/go-build,id=support-bundle-kit-go-build-${MK_REPO_ID} \
    ./scripts/build


FROM scratch AS build-output
COPY --from=build /go/src/github.com/rancher/support-bundle-kit/bin/ /bin/

# ---- validate ----
FROM base AS validate
ARG MK_REPO_ID

RUN --mount=type=cache,target=/go/pkg/mod,id=support-bundle-kit-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/github.com/rancher/support-bundle-kit/.cache/go-build,id=support-bundle-kit-go-build-${MK_REPO_ID} \
    ./scripts/validate
   
# ---- test ----
FROM base AS test
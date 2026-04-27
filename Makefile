TARGETS := $(shell ls scripts)

# renovate: datasource=github-releases depName=rancher/dapper
DAPPER_VERSION := v0.6.0
DAPPER_SUM_x86_64 := ff6105ec0a2a973d972810a2dbdb9a6bae65031d286eae082d6779e04e4c2255
DAPPER_SUM_aarch64 := cbc133224cca7593482855d8dcdec247288ec83f0fc99fbbe0ad8423260930ff

.dapper:
	@echo Downloading dapper $(DAPPER_VERSION)
	@ARCH=$$(uname -m); \
	curl -sL "https://github.com/rancher/dapper/releases/download/$(DAPPER_VERSION)/dapper-$$(uname -s)-$${ARCH}" > .dapper.tmp; \
	EXPECTED=$$(eval echo \$${DAPPER_SUM_$${ARCH}}); \
	echo "$${EXPECTED}  .dapper.tmp" | sha256sum -c -; \
	chmod +x .dapper.tmp; \
	./.dapper.tmp -v; \
	mv .dapper.tmp .dapper

$(TARGETS): .dapper
	./.dapper $@

.DEFAULT_GOAL := default

.PHONY: $(TARGETS)

module github.com/rancher/support-bundle-kit

go 1.24.0

replace (
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.1-0.20200107013213-dc14462fd587+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e
	github.com/openshift/api => github.com/openshift/api v0.0.0-20191219222812-2987a591a72c
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20191125132246-f6563a70e19a
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	github.com/rancher/apiserver => github.com/cnrancher/apiserver v0.0.0-20210302022932-069aa785cb9f
	github.com/rancher/rancher/pkg/apis => github.com/rancher/rancher/pkg/apis v0.0.0-20211208233239-77392a65423d
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20211208233239-77392a65423d
	go.etcd.io/etcd/client/v3 => go.etcd.io/etcd/client/v3 v3.5.21
	go.etcd.io/etcd/server/v3 => go.etcd.io/etcd/server/v3 v3.5.21
	go.opentelemetry.io/contrib => go.opentelemetry.io/contrib v0.20.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.34.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.34.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20230803162519-f966b187b2e5
	google.golang.org/grpc => google.golang.org/grpc v1.71.1
	google.golang.org/protobuf => google.golang.org/protobuf v1.33.0
	gopkg.in/square/go-jose.v2 => gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.0-20220521103104-8f96da9f5d5e
	k8s.io/api => k8s.io/api v0.33.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.33.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.33.1
	k8s.io/apiserver => k8s.io/apiserver v0.33.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.33.1
	k8s.io/client-go => k8s.io/client-go v0.33.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.33.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.33.1
	k8s.io/code-generator => k8s.io/code-generator v0.29.9
	k8s.io/component-base => k8s.io/component-base v0.33.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.33.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.33.1
	k8s.io/cri-api => k8s.io/cri-api v0.33.1
	k8s.io/cri-client => k8s.io/cri-client v0.33.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.33.1
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.33.1
	k8s.io/endpointslice => k8s.io/endpointslice v0.33.1
	k8s.io/externaljwt => k8s.io/externaljwt v0.33.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.33.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.33.1
	// k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.33.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.33.1
	k8s.io/kubectl => k8s.io/kubectl v0.33.1
	k8s.io/kubelet => k8s.io/kubelet v0.33.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.29.9
	k8s.io/metrics => k8s.io/metrics v0.33.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.33.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.33.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.33.1
	kubevirt.io/api => kubevirt.io/api v0.53.1
	kubevirt.io/client-go => kubevirt.io/client-go v0.53.1
	kubevirt.io/containerized-data-importer => kubevirt.io/containerized-data-importer v1.47.0
	kubevirt.io/containerized-data-importer-api => kubevirt.io/containerized-data-importer-api v1.47.0
	kubevirt.io/kubevirt => kubevirt.io/kubevirt v0.53.1
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)

require (
	github.com/Jeffail/gabs/v2 v2.6.1
	github.com/gorilla/mux v1.8.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.35.1
	github.com/pkg/errors v0.9.1
	github.com/rancher/lasso v0.0.0-20220519004610-700f167d8324
	github.com/rancher/wrangler v1.0.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.8.1
	github.com/virtual-kubelet/node-cli v0.7.0
	github.com/virtual-kubelet/virtual-kubelet v1.6.0
	go.etcd.io/etcd/client/v3 v3.5.21
	go.etcd.io/etcd/server/v3 v3.5.21
	golang.org/x/sync v0.12.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.33.1
	k8s.io/apimachinery v0.33.1
	k8s.io/apiserver v0.33.1
	k8s.io/client-go v0.33.1
	k8s.io/kubernetes v1.33.2
	k8s.io/metrics v0.33.1
)

require (
	cel.dev/expr v0.19.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.etcd.io/etcd/client/v2 v2.305.21 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.21 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.21 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/grpc v1.70.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.24.0 // indirect
	k8s.io/cloud-provider v0.0.0 // indirect
	k8s.io/cluster-bootstrap v0.0.0 // indirect
	k8s.io/code-generator v0.33.1 // indirect
	k8s.io/component-base v0.33.1 // indirect
	k8s.io/component-helpers v0.33.1 // indirect
	k8s.io/controller-manager v0.33.1 // indirect
	k8s.io/csi-translation-lib v0.0.0 // indirect
	k8s.io/dynamic-resource-allocation v0.0.0 // indirect
	k8s.io/endpointslice v0.0.0 // indirect
	k8s.io/externaljwt v0.0.0 // indirect
	k8s.io/gengo v0.0.0-20230829151522-9cce18d56c01 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kms v0.33.1 // indirect
	k8s.io/kube-aggregator v0.24.0 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	k8s.io/kubelet v0.33.1 // indirect
	k8s.io/mount-utils v0.0.0 // indirect
	k8s.io/pod-security-admission v0.0.0 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.31.2 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/go-oidc v2.3.0+incompatible // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cel-go v0.23.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/selinux v1.11.1 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_golang v1.22.0
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/testify v1.10.0
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20220101234140-673ab2c3ae75 // indirect
	github.com/xiang90/probing v0.0.0-20221125231312-a49e3df8f510 // indirect
	go.etcd.io/bbolt v1.3.11 // indirect
	go.etcd.io/etcd/api/v3 v3.5.21 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.21 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/genproto v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect

)

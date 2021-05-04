BIN := k8s-policies-controller
CRD_OPTIONS ?= "crd:trivialVersions=true"
PKG := github.com/nuxeo/k8s-policies-controller
ARCH ?= amd64
APP ?= k8s-policies-controller
NAMESPACE ?= default
RELEASE_NAME ?= k8s-policies-controller
KO_DOCKER_REPO = registry.softonic.io/k8s-policies-controller
REPOSITORY ?= gcr.io/build-jx-prod/library
VERSION ?= "$(shell git describe --tags | sed 's/^v//')"
VERSION_PKG ?= $(PKG)/pkg/version
VERSION_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LD_FLAGS := -X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).buildDate=$(VERSION_DATE)

IMAGE := $(BIN)

BUILD_IMAGE ?= golang:1.14-buster

controller-gen.bin := $(shell which controller-gen)
controller-gen.bin := $(if $(controller-gen.bin),$(controller-gen.bin),$(GOPATH)/bin/controller-gen)

jx-cli.bin := $(shell which jx-cli)
jx-cli.bin := $(if $(jx-cli.bin),$(jx-cli.bin),/usr/local/bin/jx-cli)

kustomize.bin := $(shell which kustomize)
kustomize.bin := $(if $(kustomize.bin),$(kustomize.bin),/usr/local/bin/kustomize)

kubectl.bin := $(shell which kubectl)
kubectl.bin := $(if $(kubectl.bin),$(kubectl.bin),/usr/local/bin/kubectl)

kubectl-neat.bin := $(shell which kubectl-neat)
kubectl-neat.bin := $(if $(kubectl-neat.bin),$(kubectl-neat.bin),/usr/local/bin/kubectl-neat)

.ONESHELL:

kustomizes-prod: export IMAGE_GEN = "github.com/softonic/k8s-policies-controller/cmd/k8s-policies-controller"

kustomizes:  export IMAGE_GEN = $(APP):$(VERSION)


.PHONY: all
all: dev

.PHONY: build
build: generate compile

.PHONY: compile
compile:
	go mod download
	GOARCH=${ARCH} go build -ldflags "$(LD_FLAGS)"

.PHONY: test
test:
	GOARCH=${ARCH} go test -v -ldflags "$(LD_FLAGS)" ./...

.PHONY: image
image:
	docker build -t $(IMAGE):latest -f Dockerfile .

.PHONY: dev
dev: image
	kind load docker-image $(IMAGE):$(VERSION)

.PHONY: unkustomizes
unkustomizes:
	kubectl delete -f manifest.yaml || true

.PHONY: kustomizes
kustomizes: manifest
	kustomize build kustomizes | kubectl apply -f -

.PHONY: up
up: image unkustomizes kustomizes

docker-%: tags := $(REPOSITORY)/$(IMAGE):latest $(REPOSITORY)/$(IMAGE):$(VERSION)

.PHONY: docker-tag
docker-tag: script=$(docker-tag.script)
docker-tag:
	$(foreach tag,$(tags),$(script))

define docker-tag.script =
docker tag $(IMAGE):latest $(tag)
$(newline)
endef

.PHONY: docker-push
docker-push: script=$(docker-push.script)
docker-push:
	$(foreach tag,$(tags),$(script))

define docker-push.script =
docker push $(tag)
$(newline)
endef

.PHONY: version
version:
	@echo $(VERSION)

null  :=
space := $(null) #
comma := ,

define newline :=

endef

.PHONY: generate
generate: packages := gcpauth gcpworkload node meta
generate: script=$(generate.script)
generate: | $(controller-gen.bin) $(jx-cli.bin) $(kustomize.bin)
generate:
	$(foreach package,$(packages),$(script))

define generate.script =
	$(controller-gen.bin) object paths=./pkg/apis/$(package)/...
	$(controller-gen.bin) crd paths=./pkg/apis/$(package)/... output:crd:artifacts:config=kustomizes/$(package)
	$(controller-gen.bin) rbac:roleName=$(package)-controller paths=./pkg/apis/$(package)/... output:rbac:artifacts:config=kustomizes/$(package)
	$(jx-cli.bin) gitops rename --dir=kustomizes/$(package)
$(newline)
endef

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

$(GOPATH)/bin/controller-gen:
	tmpdir=$$(mktemp -d)
	cd $$tmpdir
	go mod init tmp
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
	rm -rf $$tmpdir

ifeq ('$(findstring ;,$(PATH))',';')
os := windows
else
os := $(shell uname 2>/dev/null || echo Unknown)
os := $(patsubst CYGWIN%,ygwin,$(os))
os := $(patsubst MSYS%,MSYS,$(os))
os := $(patsubst MINGW%,MSYS,$(os))
os := $(shell echo $(os) | tr '[:upper:]' '[:lower:]')
endif

/usr/local/bin/jx-cli:
	curl -sL https://github.com/jenkins-x/jx-cli/releases/download/v3.1.211/jx-cli-$(os)-amd64.tar.gz| tar xvz -C /usr/local/bin -f - jx
	mv /usr/local/bin/jx $(@)
	chmod +x $(@)

$(HOME)/.config/kustomize/plugin /usr/local/bin/kustomize:
	curl -sL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.9.2/kustomize_v3.9.2_$(os)_amd64.tar.gz | tar xvz -C /usr/local/bin -f - kustomize
	chmod +x $(@)
	mkdir -p $(HOME)/.config/kustomize/plugin

/usr/local/bin/kubectl: version := $(shell curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
/usr/local/bin/kubectl:
	curl -o $(@) -L https://storage.googleapis.com/kubernetes-release/release/$(version)/bin/$(os)/amd64/kubectl
	chmod +x $(@)

/usr/local/bin/kubectl-neat:
 	curl -sL https://github.com/itaysk/kubectl-neat/releases/download/v2.0.2/kubectl-neat_$(os).tar.gz | tar xvz -C /usr/local/bin -f - kubectl-neat
	chmod +x $(@)

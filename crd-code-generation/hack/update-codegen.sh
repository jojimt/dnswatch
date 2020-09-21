#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${PWD})

vendor/k8s.io/code-generator/generate-groups.sh all \
  github.com/jojimt/dnswatch/pkg/crd github.com/jojimt/dnswatch/pkg/crd/apis \
  dnswatch:v1alpha \
  --go-header-file ${SCRIPT_ROOT}/dnswatch/crd-code-generation/hack/custom-boilerplate.go.txt

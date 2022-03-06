#!/bin/bash

echo $IMAGE_VERSION

make docker-build docker-push IMG="quay.io/linxianer12/coredns-integration-operator:$IMAGE_VERSION"

make bundle IMG="quay.io/linxianer12/coredns-integration-operator:$IMAGE_VERSION"

make bundle-build bundle-push BUNDLE_IMG="quay.io/linxianer12/coredns-integration-bundle:$IMAGE_VERSION"

operator-sdk cleanup coredns-integration-operator

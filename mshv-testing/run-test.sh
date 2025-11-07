#!/bin/bash

export DOCKER_PREFIX=ghcr.io/kaizentm/kubevirt
export DOCKER_TAG=1.6.0-l1vh.196
export KUBEVIRT_PROVIDER=external
export KUBEVIRT_WITH_ETC_IN_MEMORY=true
export KUBEVIRT_FUNC_TEST_TIMEOUT="8h"
export KUBEVIRT_FUNC_TEST_SUITE_ARGS="--ginkgo.v"
export KUBECONFIG=/home/ubuntu/git/ARO/kubevirt/mshv-testing/kubeconfig-dev

export TARGET=k8s-1.32-sig-compute
export KUBEVIRT_E2E_PARALLEL=true

#TODO Check if these are needed
#make cluster-up
#make cluster-deploy

#./automation/test.sh

make cluster-up

export FUNC_TEST_LABEL_FILTER='--label-filter=(!flake-check)&&((sig-compute && !(GPU,VGPU,sig-compute-migrations,sig-storage) && !(SEV, SEVES))&&(!Windows)&&(!Sysprep)&&(!requires-s390x)&&(!requires-arm64)&&(!RequiresVolumeExpansion)&&!(single-replica)&&(!requireHugepages2Mi)&&(!requireHugepages1Gi)&&(!SwapTest))'

export KUBEVIRT_E2E_PARALLEL=false # DOnt run parallel since were only running 1 test at a time
export KUBEVIRT_FUNC_TEST_SUITE_ARGS="${KUBEVIRT_FUNC_TEST_SUITE_ARGS} --ginkgo.focus=\[sig\-compute\]Migration\srecovery\sshould\ssuccessfully\sdefer\sa\smigration\sfailure"
make functest 


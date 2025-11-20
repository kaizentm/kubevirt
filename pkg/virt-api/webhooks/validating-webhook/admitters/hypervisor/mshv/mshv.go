package mshv_validator

import (
	base_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor/base"
)

type MshvValidator struct {
	*base_validator.BaseValidator
}

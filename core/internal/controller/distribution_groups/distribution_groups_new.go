package distribution_groups

import (
	"billionmail-core/api/distribution_groups"
)

type ControllerV1 struct{}

func NewV1() distribution_groups.IDistributionGroupsV1 {
	return &ControllerV1{}
}

package distribution_groups

import (
	"context"

	v1 "billionmail-core/api/distribution_groups/v1"
)

type IDistributionGroupsV1 interface {
	CreateGroup(ctx context.Context, req *v1.CreateGroupReq) (res *v1.CreateGroupRes, err error)
}

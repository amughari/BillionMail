package distribution_groups

import (
	"context"
	"fmt"

	v1 "billionmail-core/api/distribution_groups/v1"
	groupService "billionmail-core/internal/service/distribution_groups"
	"billionmail-core/internal/service/rbac"

	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) CreateGroup(
	ctx context.Context,
	req *v1.CreateGroupReq,
) (res *v1.CreateGroupRes, err error) {
	res = &v1.CreateGroupRes{}

	// Get authenticated account ID from JWT context.
	adminID := rbac.GetCurrentAccountId(ctx)
	if adminID <= 0 {
		res.Code = 401
		res.SetError(fmt.Errorf("authentication required"))
		return res, nil
	}

	// Only global administrators can manage distribution groups.
	isAdmin, err := rbac.Account().IsAdmin(ctx, adminID)
	if err != nil {
		res.Code = 500
		res.SetError(fmt.Errorf("failed to verify administrator permissions"))
		return res, nil
	}

	if !isAdmin {
		res.Code = 403
		res.SetError(fmt.Errorf("administrator privileges required"))
		return res, nil
	}

	clientIP := g.RequestFromCtx(ctx).GetClientIp()

	groupID, err := groupService.CreateGroup(
		ctx,
		adminID,
		req.Name,
		req.Email,
		req.Description,
		clientIP,
	)
	if err != nil {
		res.Code = 400
		res.SetError(err)
		return res, nil
	}

	res.Data.ID = groupID
	res.Data.Name = req.Name
	res.Data.Email = req.Email
	res.SetSuccess("Distribution group created successfully")

	return res, nil
}

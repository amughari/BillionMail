package v1

import (
	"billionmail-core/utility/types/api_v1"

	"github.com/gogf/gf/v2/frame/g"
)

type CreateGroupReq struct {
	g.Meta `path:"/distribution-groups/create" method:"post" tags:"Distribution Groups" summary:"Create distribution group"`

	Name        string `json:"name" v:"required" dc:"Group name"`
	Email       string `json:"email" v:"required|email" dc:"Group email address"`
	Description string `json:"description" dc:"Group description"`
}

type CreateGroupRes struct {
	api_v1.StandardRes
	Data struct {
		ID    int64  `json:"id" dc:"Created group ID"`
		Name  string `json:"name" dc:"Group name"`
		Email string `json:"email" dc:"Group email address"`
	} `json:"data"`
}

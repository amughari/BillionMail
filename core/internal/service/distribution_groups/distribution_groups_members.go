package distribution_groups

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AddMailboxMember adds an existing active mailbox to a distribution group.
func AddMailboxMember(
	ctx context.Context,
	db gdb.DB,
	adminID, groupID int64,
	address, ip string,
) error {
	address = strings.ToLower(strings.TrimSpace(address))
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Address != address {
		return fmt.Errorf("invalid mailbox address")
	}

	return db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		groupCount, err := tx.Model("bm_distribution_groups").
			Ctx(ctx).
			Where("id", groupID).
			Where("active", 1).
			Count()
		if err != nil {
			return err
		}
		if groupCount == 0 {
			return fmt.Errorf("distribution group does not exist or is inactive")
		}

		mailboxCount, err := tx.Model("mailbox").
			Ctx(ctx).
			Where("LOWER(username) = ?", address).
			Where("active", 1).
			Count()
		if err != nil {
			return err
		}
		if mailboxCount == 0 {
			return fmt.Errorf("active mailbox does not exist")
		}

		memberCount, err := tx.Model("bm_distribution_members").
			Ctx(ctx).
			Where("group_id", groupID).
			Where("member_type", "mailbox").
			Where("LOWER(mailbox_address) = ?", address).
			Count()
		if err != nil {
			return err
		}
		if memberCount > 0 {
			return fmt.Errorf("mailbox is already a member of this group")
		}

		now := time.Now().Unix()
		_, err = tx.Model("bm_distribution_members").Ctx(ctx).Data(g.Map{
			"group_id":        groupID,
			"member_type":     "mailbox",
			"mailbox_address": address,
			"created_at":      now,
		}).Insert()
		if err != nil {
			return err
		}

		return recordMemberAudit(ctx, tx, adminID, groupID, "add_mailbox_member",
			g.Map{"mailbox": address}, ip, now)
	})
}

// AddGroupMember adds another distribution group as a nested member.
// It rejects self-membership and circular group references.
func AddGroupMember(
	ctx context.Context,
	db gdb.DB,
	adminID, parentGroupID, childGroupID int64,
	ip string,
) error {
	if parentGroupID == childGroupID {
		return fmt.Errorf("a group cannot contain itself")
	}

	return db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, id := range []int64{parentGroupID, childGroupID} {
			count, err := tx.Model("bm_distribution_groups").
				Ctx(ctx).
				Where("id", id).
				Where("active", 1).
				Count()
			if err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("distribution group %d does not exist or is inactive", id)
			}
		}

		// Starting at the child, walk its descendants.
		// If the parent is reachable, adding this edge would create a cycle.
		visited := map[int64]bool{}
		queue := []int64{childGroupID}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if current == parentGroupID {
				return fmt.Errorf("adding this member would create a circular group reference")
			}
			if visited[current] {
				continue
			}
			visited[current] = true

			children, err := tx.Model("bm_distribution_members").
				Ctx(ctx).
				Fields("child_group_id").
				Where("group_id", current).
				Where("member_type", "group").
				Array()
			if err != nil {
				return err
			}

			for _, value := range children {
				nextID := value.Int64()
				if !visited[nextID] {
					queue = append(queue, nextID)
				}
			}
		}

		existing, err := tx.Model("bm_distribution_members").
			Ctx(ctx).
			Where("group_id", parentGroupID).
			Where("member_type", "group").
			Where("child_group_id", childGroupID).
			Count()
		if err != nil {
			return err
		}
		if existing > 0 {
			return fmt.Errorf("child group is already a member")
		}

		now := time.Now().Unix()
		_, err = tx.Model("bm_distribution_members").Ctx(ctx).Data(g.Map{
			"group_id":       parentGroupID,
			"member_type":    "group",
			"child_group_id": childGroupID,
			"created_at":     now,
		}).Insert()
		if err != nil {
			return err
		}

		return recordMemberAudit(ctx, tx, adminID, parentGroupID, "add_child_group",
			g.Map{"child_group_id": childGroupID}, ip, now)
	})
}

func recordMemberAudit(
	ctx context.Context,
	tx gdb.TX,
	adminID, groupID int64,
	action string,
	details g.Map,
	ip string,
	now int64,
) error {
	data, err := json.Marshal(details)
	if err != nil {
		return err
	}

	_, err = tx.Model("bm_distribution_audit").Ctx(ctx).Data(g.Map{
		"admin_id":   adminID,
		"group_id":   groupID,
		"action":     action,
		"details":    string(data),
		"ip":         ip,
		"created_at": now,
	}).Insert()

	return err
}

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

// CreateGroup creates a distribution group and records the operation.
func CreateGroup(
	ctx context.Context,
	adminID int64,
	name string,
	email string,
	description string,
	ip string,
) (int64, error) {
	return CreateGroupWithDB(ctx, g.DB(), adminID, name, email, description, ip)
}

// CreateGroupWithDB creates a distribution group using the supplied database connection.
func CreateGroupWithDB(
	ctx context.Context,
	db gdb.DB,
	adminID int64,
	name string,
	email string,
	description string,
	ip string,
) (int64, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	description = strings.TrimSpace(description)

	if name == "" {
		return 0, fmt.Errorf("group name is required")
	}

	if email == "" {
		return 0, fmt.Errorf("group email is required")
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return 0, fmt.Errorf("invalid group email address")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, fmt.Errorf("invalid group email address")
	}

	now := time.Now().Unix()
	var groupID int64

	err = db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Verify that the domain exists and is active.
		domainCount, err := tx.Model("domain").
			Ctx(ctx).
			Where("domain", parts[1]).
			Where("active", 1).
			Count()
		if err != nil {
			return err
		}
		if domainCount == 0 {
			return fmt.Errorf("email domain does not exist or is inactive")
		}

		// Prevent conflicts with existing mailboxes.
		mailboxCount, err := tx.Model("mailbox").
			Ctx(ctx).
			Where("LOWER(username) = ?", email).
			Count()
		if err != nil {
			return err
		}
		if mailboxCount > 0 {
			return fmt.Errorf("email address already belongs to a mailbox")
		}

		// Prevent conflicts with existing aliases.
		aliasCount, err := tx.Model("alias").
			Ctx(ctx).
			Where("LOWER(address) = ?", email).
			Count()
		if err != nil {
			return err
		}
		if aliasCount > 0 {
			return fmt.Errorf("email address already belongs to an alias")
		}

		// Prevent duplicate distribution group addresses.
		groupCount, err := tx.Model("bm_distribution_groups").
			Ctx(ctx).
			Where("LOWER(email) = ?", email).
			Count()
		if err != nil {
			return err
		}
		if groupCount > 0 {
			return fmt.Errorf("distribution group email already exists")
		}

		// Create the group.
		groupID, err = tx.Model("bm_distribution_groups").
			Ctx(ctx).
			Data(g.Map{
				"name":        name,
				"email":       email,
				"description": description,
				"active":      1,
				"created_by":  adminID,
				"created_at":  now,
				"updated_at":  now,
			}).
			InsertAndGetId()
		if err != nil {
			return err
		}

		// Record the creation in the audit log.
		details, err := json.Marshal(g.Map{
			"name":  name,
			"email": email,
		})
		if err != nil {
			return err
		}

		_, err = tx.Model("bm_distribution_audit").
			Ctx(ctx).
			Data(g.Map{
				"admin_id":   adminID,
				"group_id":   groupID,
				"action":     "create",
				"details":    string(details),
				"ip":         ip,
				"created_at": now,
			}).
			Insert()

		return err
	})

	if err != nil {
		return 0, err
	}

	return groupID, nil
}

package database_initialization

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	registerHandler(func() {
		ctx := context.Background()

		sqlList := []string{
			// Distribution groups
			`CREATE TABLE IF NOT EXISTS bm_distribution_groups (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				email VARCHAR(255) NOT NULL,
				description TEXT NOT NULL DEFAULT '',
				active SMALLINT NOT NULL DEFAULT 1
					CHECK (active IN (0, 1)),
				created_by BIGINT NOT NULL DEFAULT 0,
				created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
				updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
			)`,

			`CREATE UNIQUE INDEX IF NOT EXISTS idx_bm_dgroups_email
				ON bm_distribution_groups (LOWER(email))`,

			`CREATE INDEX IF NOT EXISTS idx_bm_dgroups_active
				ON bm_distribution_groups (active)`,

			// Distribution group members
			`CREATE TABLE IF NOT EXISTS bm_distribution_members (
				id BIGSERIAL PRIMARY KEY,
				group_id BIGINT NOT NULL
					REFERENCES bm_distribution_groups(id) ON DELETE CASCADE,
				member_type VARCHAR(16) NOT NULL
					CHECK (member_type IN ('mailbox', 'group')),
				mailbox_address VARCHAR(255),
				child_group_id BIGINT
					REFERENCES bm_distribution_groups(id) ON DELETE RESTRICT,
				created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),

				CONSTRAINT chk_bm_dgroup_member_target CHECK (
					(member_type = 'mailbox'
						AND mailbox_address IS NOT NULL
						AND child_group_id IS NULL)
					OR
					(member_type = 'group'
						AND mailbox_address IS NULL
						AND child_group_id IS NOT NULL)
				)
			)`,

			`CREATE UNIQUE INDEX IF NOT EXISTS idx_bm_dmembers_mailbox
				ON bm_distribution_members (group_id, LOWER(mailbox_address))
				WHERE member_type = 'mailbox'`,

			`CREATE UNIQUE INDEX IF NOT EXISTS idx_bm_dmembers_child_group
				ON bm_distribution_members (group_id, child_group_id)
				WHERE member_type = 'group'`,

			`CREATE INDEX IF NOT EXISTS idx_bm_dmembers_group
				ON bm_distribution_members (group_id)`,

			`CREATE INDEX IF NOT EXISTS idx_bm_dmembers_child
				ON bm_distribution_members (child_group_id)
				WHERE child_group_id IS NOT NULL`,

			// Distribution group audit log
			`CREATE TABLE IF NOT EXISTS bm_distribution_audit (
				id BIGSERIAL PRIMARY KEY,
				admin_id BIGINT NOT NULL DEFAULT 0,
				group_id BIGINT,
				action VARCHAR(32) NOT NULL,
				details JSONB NOT NULL DEFAULT '{}'::jsonb,
				ip VARCHAR(64) NOT NULL DEFAULT '',
				created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
			)`,

			`CREATE INDEX IF NOT EXISTS idx_bm_dgaudit_group
				ON bm_distribution_audit (group_id)`,

			`CREATE INDEX IF NOT EXISTS idx_bm_dgaudit_admin
				ON bm_distribution_audit (admin_id)`,

			`CREATE INDEX IF NOT EXISTS idx_bm_dgaudit_created
				ON bm_distribution_audit (created_at)`,
		}

		for _, sql := range sqlList {
			if _, err := g.DB().Exec(ctx, sql); err != nil {
				g.Log().Error(ctx,
					"Failed to initialize distribution groups database:",
					err,
				)
				return
			}
		}

		g.Log().Info(ctx, "Distribution groups database initialized successfully")
	})
}

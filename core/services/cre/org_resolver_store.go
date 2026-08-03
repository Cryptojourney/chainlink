package cre

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
)

// orgResolverCacheTable is the durable owner->orgID mapping table backing the
// OrgResolver cache. See migration 0305_org_resolver_cache.sql.
const orgResolverCacheTable = "cre.org_resolver_cache"

// orgResolverStore is a Postgres-backed implementation of orgresolver.CacheStore.
type orgResolverStore struct {
	ds sqlutil.DataSource
}

// NewOrgResolverStore creates a durable cache store for the OrgResolver.
func NewOrgResolverStore(ds sqlutil.DataSource) *orgResolverStore {
	return &orgResolverStore{ds: ds}
}

// GetOrg returns the cached orgID for owner, or sql.ErrNoRows if absent.
func (s *orgResolverStore) GetOrg(ctx context.Context, owner string) (string, error) {
	const q = `SELECT org_id FROM ` + orgResolverCacheTable + ` WHERE workflow_owner = $1`
	var orgID string
	if err := s.ds.GetContext(ctx, &orgID, q, owner); err != nil {
		return "", fmt.Errorf("failed to get cached org for owner %s: %w", owner, err)
	}
	return orgID, nil
}

// UpsertOrg stores or updates the owner->orgID mapping.
func (s *orgResolverStore) UpsertOrg(ctx context.Context, owner, orgID string) error {
	const q = `
INSERT INTO ` + orgResolverCacheTable + ` (workflow_owner, org_id, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (workflow_owner) DO UPDATE SET org_id = EXCLUDED.org_id, updated_at = NOW()`
	if _, err := s.ds.ExecContext(ctx, q, owner, orgID); err != nil {
		return fmt.Errorf("failed to upsert org for owner %s: %w", owner, err)
	}
	return nil
}

var _ interface {
	GetOrg(ctx context.Context, owner string) (string, error)
	UpsertOrg(ctx context.Context, owner, orgID string) error
} = (*orgResolverStore)(nil)

package tenantadm

import (
	"context"
)

type Tenant struct {
	ID string `json:"id"`
}

type Client interface {
	GetTenantByToken(ctx context.Context, tenantToken string) (*Tenant, error)
}

package api

import (
	"context"
	"net/http"
)

// DataAccessPolicy controls default EdgeKV data access restrictions.
type DataAccessPolicy struct {
	// RestrictDataAccess enables restricted database access.
	RestrictDataAccess bool `json:"restrictDataAccess"`
	// AllowNamespacePolicyOverride permits namespaces to override the default policy.
	AllowNamespacePolicyOverride bool `json:"allowNamespacePolicyOverride"`
}

// DatabaseStatus describes EdgeKV initialization and network status.
type DatabaseStatus struct {
	// AccountStatus is the database initialization state.
	AccountStatus string `json:"accountStatus,omitempty"`
	// CPCode identifies the database traffic reporting code.
	CPCode string `json:"cpcode,omitempty"`
	// ProductionStatus is the production network state.
	ProductionStatus string `json:"productionStatus,omitempty"`
	// StagingStatus is the staging network state.
	StagingStatus string `json:"stagingStatus,omitempty"`
	// DataAccessPolicy is the database's default access policy when returned.
	DataAccessPolicy *DataAccessPolicy `json:"dataAccessPolicy,omitempty"`
}

// InitializeDatabase starts EdgeKV database initialization.
func (client *Client) InitializeDatabase(ctx context.Context, policy *DataAccessPolicy) (*DatabaseStatus, error) {
	var body any
	if policy != nil {
		body = map[string]any{"dataAccessPolicy": policy}
	}
	result := &DatabaseStatus{}
	if err := client.do(ctx, http.MethodPut, "/edgekv/v1/initialize", body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetDatabaseStatus returns the EdgeKV database initialization status.
func (client *Client) GetDatabaseStatus(ctx context.Context) (*DatabaseStatus, error) {
	result := &DatabaseStatus{}
	if err := client.do(ctx, http.MethodGet, "/edgekv/v1/initialize", nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateDatabasePolicy replaces the default EdgeKV data access policy.
func (client *Client) UpdateDatabasePolicy(ctx context.Context, policy DataAccessPolicy) (*DatabaseStatus, error) {
	result := &DatabaseStatus{}
	if err := client.do(ctx, http.MethodPut, "/edgekv/v1/auth/database", policy, result); err != nil {
		return nil, err
	}
	return result, nil
}

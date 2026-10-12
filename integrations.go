package nusii

import (
	"context"
	"net/http"
)

// IntegrationsService connects the account to CRMs.
type IntegrationsService struct{ n *Nusii }

// OnePageCRMInstallationParams are the OnePageCRM credentials to connect with.
type OnePageCRMInstallationParams struct {
	Name      Opt[string] `json:"name,omitzero"`
	APIKey    string      `json:"api_key"`
	APISecret string      `json:"api_secret"`
}

// LessAnnoyingCRMInstallationParams are the Less Annoying CRM credentials to connect with.
type LessAnnoyingCRMInstallationParams struct {
	Name     Opt[string] `json:"name,omitzero"`
	Token    string      `json:"token"`
	UserCode string      `json:"user_code"`
	UserID   Opt[string] `json:"user_id,omitzero"`
}

// CreateOnePageCRMInstallation connects the account to OnePageCRM.
func (s *IntegrationsService) CreateOnePageCRMInstallation(ctx context.Context, params OnePageCRMInstallationParams, opts ...Option) (*IntegrationInstallation, error) {
	body := map[string]any{"one_page_crm_installation": params}
	return resource[IntegrationInstallation](ctx, s.n, http.MethodPost, "/integrations/one_page_crm_installations", body, opts)
}

// CreateLessAnnoyingCRMInstallation connects the account to Less Annoying CRM.
func (s *IntegrationsService) CreateLessAnnoyingCRMInstallation(ctx context.Context, params LessAnnoyingCRMInstallationParams, opts ...Option) (*IntegrationInstallation, error) {
	body := map[string]any{"less_annoying_crm_installation": params}
	return resource[IntegrationInstallation](ctx, s.n, http.MethodPost, "/integrations/less_annoying_crm_installations", body, opts)
}

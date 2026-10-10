package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/config"
	"github.com/danielgtaylor/huma/v2"
)

type rootGeneralSettings struct {
	RegistrationEnabled bool `json:"registration_enabled"`
}

func init() { AddRouteRegistrar(RegisterRootSettingsRoutes) }

func RegisterRootSettingsRoutes(api huma.API) {
	tags := []string{"root"}
	Register(api, huma.Operation{OperationID: "root-settings-general-read", Summary: "Read root settings", Description: "Returns the instance settings available to an authenticated instance administrator.", Method: http.MethodGet, Path: "/root/settings/general", Tags: tags}, rootSettingsGeneralRead)
	Register(api, huma.Operation{OperationID: "root-settings-general-update", Summary: "Update root settings", Description: "Persists whether public registration is open for the instance.", Method: http.MethodPut, Path: "/root/settings/general", Tags: tags}, rootSettingsGeneralUpdate)
}

func rootSettingsGeneralRead(ctx context.Context, _ *struct{}) (*singleBody[rootGeneralSettings], error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	return &singleBody[rootGeneralSettings]{Body: &rootGeneralSettings{RegistrationEnabled: config.ServiceEnableRegistration.GetBool()}}, nil
}

func rootSettingsGeneralUpdate(ctx context.Context, in *struct{ Body rootGeneralSettings }) (*singleBody[rootGeneralSettings], error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	if err := config.SetRegistrationEnabled(in.Body.RegistrationEnabled); err != nil {
		return nil, huma.Error500InternalServerError("could not persist registration setting")
	}
	return &singleBody[rootGeneralSettings]{Body: &in.Body}, nil
}

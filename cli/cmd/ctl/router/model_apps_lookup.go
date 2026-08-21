package router

import (
	"context"
	"strings"
)

// The model applications the Market publishes, read for one purpose only:
// naming the provider of an application that is still installing.
//
// GET /console/api/model-apps
//
// There is no verb for this. Installing, cloning, upgrading and removing a model
// application are the Market's own, so `olares-cli market` is where they live;
// what remains here is the lookup, because a freshly started install is absent
// from both the provider list and the aggregate model list while being exactly
// what someone is most likely to name.

// modelApp is the part of a row this lookup needs: the application's name, and
// the provider Router created for the copy installed here. ProviderID is empty
// unless a copy is installed.
type modelApp struct {
	AppName string `json:"app_name"`
	Install struct {
		ProviderID string `json:"provider_id"`
	} `json:"install"`
}

// providerIDFromModelApps finds a provider id by application name among the
// model applications the Market publishes, which is where a row hidden from
// every other list still appears.
//
// Empty when the name is unknown there, when the copy holding it belongs to
// another source, or when there is no Market to ask — this is a fallback, and
// its failure has to read as "not found" rather than replace the caller's error
// with one about a different route.
func providerIDFromModelApps(ctx context.Context, pc *preparedClient, appName string) string {
	apps, err := collection[modelApp](ctx, pc, epModelApps)
	if err != nil {
		return ""
	}
	for i := range apps {
		if strings.EqualFold(apps[i].AppName, appName) && apps[i].Install.ProviderID != "" {
			return apps[i].Install.ProviderID
		}
	}
	return ""
}

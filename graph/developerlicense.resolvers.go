package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"

	"github.com/DIMO-Network/identity-api/graph/model"
)

// Signers is the resolver for the signers field.
func (r *developerLicenseResolver) Signers(ctx context.Context, obj *model.DeveloperLicense, first *int, after *string, last *int, before *string) (*model.SignerConnection, error) {
	return r.developerLicense.GetSignersForLicense(ctx, obj, first, after, last, before)
}

// RedirectURIs is the resolver for the redirectURIs field.
func (r *developerLicenseResolver) RedirectURIs(ctx context.Context, obj *model.DeveloperLicense, first *int, after *string, last *int, before *string) (*model.RedirectURIConnection, error) {
	return r.developerLicense.GetRedirectURIsForLicense(ctx, obj, first, after, last, before)
}

// DeveloperLicenses is the resolver for the developerLicenses field.
func (r *queryResolver) DeveloperLicenses(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.DeveloperLicenseFilterBy) (*model.DeveloperLicenseConnection, error) {
	return r.developerLicense.GetDeveloperLicenses(ctx, first, after, last, before, filterBy)
}

// DeveloperLicense is the resolver for the developerLicense field.
func (r *queryResolver) DeveloperLicense(ctx context.Context, by model.DeveloperLicenseBy) (*model.DeveloperLicense, error) {
	return r.developerLicense.GetLicense(ctx, by)
}

// DeveloperLicense returns DeveloperLicenseResolver implementation.
func (r *Resolver) DeveloperLicense() DeveloperLicenseResolver { return &developerLicenseResolver{r} }

type developerLicenseResolver struct{ *Resolver }

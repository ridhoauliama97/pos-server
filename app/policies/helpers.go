package policies

import (
	"context"
	"errors"

	accessimpl "github.com/goravel/framework/auth/access"
	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/http"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/models"
)

// currentUser returns the authenticated user bound to the request context.
func currentUser(ctx context.Context) (*models.User, error) {
	httpContext, ok := ctx.(http.Context)
	if !ok {
		return nil, errors.New("missing http context")
	}

	var user models.User
	if err := facades.Auth(httpContext).User(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// allowRoles grants the ability when the authenticated user has one of the given roles.
func allowRoles(roles ...string) func(ctx context.Context, arguments map[string]any) access.Response {
	return func(ctx context.Context, arguments map[string]any) access.Response {
		user, err := currentUser(ctx)
		if err != nil {
			return accessimpl.NewDenyResponse("unauthenticated")
		}

		for _, role := range roles {
			if user.Role == role {
				return accessimpl.NewAllowResponse()
			}
		}

		return accessimpl.NewDenyResponse("this action is unauthorized")
	}
}

// and grants the ability only when every given policy allows it.
func and(policies ...func(ctx context.Context, arguments map[string]any) access.Response) func(ctx context.Context, arguments map[string]any) access.Response {
	return func(ctx context.Context, arguments map[string]any) access.Response {
		for _, policy := range policies {
			if response := policy(ctx, arguments); !response.Allowed() {
				return response
			}
		}

		return accessimpl.NewAllowResponse()
	}
}

// AllowOwners grants the given ability to owner role users.
func AllowOwners() func(ctx context.Context, arguments map[string]any) access.Response {
	return allowRoles(models.RoleOwner)
}

// DenyOwnerRole denies requests that try to create or promote a user to the owner role.
func DenyOwnerRole() func(ctx context.Context, arguments map[string]any) access.Response {
	return func(ctx context.Context, arguments map[string]any) access.Response {
		if role, ok := arguments["role"].(string); ok && role == models.RoleOwner {
			return accessimpl.NewDenyResponse("only an owner can manage owner users")
		}

		return accessimpl.NewAllowResponse()
	}
}

// Register defines every role ability on the application gate.
func Register(gate access.Gate) {
	gate.Before(func(ctx context.Context, ability string, arguments map[string]any) access.Response {
		user, err := currentUser(ctx)
		if err == nil && user.Role == models.RoleOwner {
			return accessimpl.NewAllowResponse()
		}

		return nil
	})

	RegisterProduct(gate)
	RegisterCategory(gate)
	RegisterStock(gate)
	RegisterCustomer(gate)
	RegisterTransaction(gate)
	RegisterReport(gate)
	RegisterUser(gate)
	RegisterOutlet(gate)
}

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

type Mware struct{}
type contextKey struct {
	name string
}

var authCtxKey = &contextKey{"auth"}

type User struct {
	Roles       []string
	Permissions []string
}

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("auth")

			// TODO - Do checks if user is allowed using DB
			ctx := context.WithValue(r.Context(), authCtxKey, auth)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext finds the user from the context. REQUIRES Middleware to have run.
func GetUserFromContext(ctx context.Context) *User {
	authData := ctx.Value(authCtxKey).(string)

	user := &User{}
	json.Unmarshal([]byte(authData), &user)

	return user
}

/* // Lifecycle hooks of gqlGen


type FedAuth struct {
	DisableColor bool
}

var _ interface {
	graphql.HandlerExtension
} = &FedAuth{}

func (a FedAuth) ExtensionName() string {
	return "DimoAuth"
}

func (a *FedAuth) Validate(schema graphql.ExecutableSchema) error {

	return nil
}

func (a *FedAuth) MutateOperationParameters(ctx context.Context, request *graphql.RawParams) *gqlerror.Error {
	auth := request.Headers.Get("auth")
	if auth == "" {
		return gqlerror.Errorf("Unauthorized: authorization header is missing from request")
	}

	// log.Println(request.Headers.Get("auth"), "111MutateOperationParameters")
	return nil
}

func (a *FedAuth) MutateOperationContext(ctx context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	auth := rc.Headers.Get("auth")
	if auth == "" {
		return gqlerror.Errorf("Unauthorized: authorization header is missing from request")
	}

	ctx = context.WithValue(ctx, "auth", auth)

	// log.Println(auth, "22222MutateOperationContext")
	return nil
}
*/

// Package auth provides authentication for the GraphQL API.
package auth

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// PublicOperations lists GraphQL operations that don't require authentication.
// All other operations require a valid JWT token.
var PublicOperations = map[string]bool{
	"login":              true,// Subscription - auth handled via WebSocket connectionParams
	"IntrospectionQuery": true, // For GraphQL tooling
	"__schema":           true, // For GraphQL introspection
	"__type":             true, // For GraphQL introspection
}

// AuthExtension is a gqlgen extension that enforces authentication
// on all operations except those in PublicOperations.
type AuthExtension struct{}

var _ interface {
	graphql.OperationInterceptor
	graphql.HandlerExtension
} = AuthExtension{}

// ExtensionName returns the name of this extension.
func (AuthExtension) ExtensionName() string {
	return "AuthExtension"
}

// Validate is called when adding the extension to the server.
func (AuthExtension) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

// InterceptOperation is called before each GraphQL operation.
// It checks if the operation requires authentication and validates the user.
func (AuthExtension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	opCtx := graphql.GetOperationContext(ctx)

	// Check if this is a public operation
	if isPublicOperation(opCtx) {
		return next(ctx)
	}

	// Require authentication for all other operations
	_, ok := GetUserFromContext(ctx)
	if !ok {
		return func(ctx context.Context) *graphql.Response {
			return &graphql.Response{
				Errors: gqlerror.List{
					&gqlerror.Error{
						Message: "authentication required",
						Extensions: map[string]interface{}{
							"code": "UNAUTHENTICATED",
						},
					},
				},
			}
		}
	}

	return next(ctx)
}

// isPublicOperation checks if the current operation is in the public list.
func isPublicOperation(opCtx *graphql.OperationContext) bool {
	// Check by operation name first (named operations like "IntrospectionQuery")
	if opCtx.OperationName != "" {
		if PublicOperations[opCtx.OperationName] {
			return true
		}
	}

	// Check the actual fields being queried/mutated
	if opCtx.Operation == nil {
		return false
	}

	// Check each field in the selection set
	for _, sel := range opCtx.Operation.SelectionSet {
		if field, ok := sel.(*ast.Field); ok {
			if PublicOperations[field.Name] {
				return true
			}
		}
	}

	return false
}
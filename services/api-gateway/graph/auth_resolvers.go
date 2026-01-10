package graph

import (
	"context"
	"fmt"

	"github.com/yourusername/iot-platform/services/api-gateway/auth"
	"github.com/yourusername/iot-platform/services/api-gateway/graph/model"
	userpb "github.com/yourusername/iot-platform/shared/proto/user"
)

// Helper functions to convert between Protobuf and GraphQL types for users

func protoToGraphQLUser(u *userpb.User) *model.User {
	if u == nil {
		return nil
	}

	return &model.User{
		ID:        u.Id,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: int(u.CreatedAt),
		LastLogin: intPtr(int(u.LastLogin)),
		IsActive:  u.IsActive,
	}
}

// Mutation resolvers for authentication

func (r *mutationResolver) RegisterImpl(ctx context.Context, input model.RegisterInput) (*model.AuthPayload, error) {
	// Prepare register request
	req := &userpb.RegisterRequest{
		Email:    input.Email,
		Password: input.Password,
		Name:     input.Name,
	}

	// Set role if provided, otherwise defaults to "user" in user-service
	if input.Role != nil {
		req.Role = *input.Role
	}

	// Call User Service via gRPC
	resp, err := r.UserClient.Register(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// Generate JWT token
	token, err := r.JWTManager.GenerateToken(
		resp.User.Id,
		resp.User.Email,
		resp.User.Name,
		resp.User.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.AuthPayload{
		Token: token,
		User:  protoToGraphQLUser(resp.User),
	}, nil
}

func (r *mutationResolver) LoginImpl(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	// Prepare authenticate request
	req := &userpb.AuthenticateRequest{
		Email:    input.Email,
		Password: input.Password,
	}

	// Call User Service via gRPC
	resp, err := r.UserClient.Authenticate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Check if authentication was successful
	if !resp.Success {
		return nil, fmt.Errorf("authentication failed: %s", resp.Message)
	}

	// Generate JWT token
	token, err := r.JWTManager.GenerateToken(
		resp.User.Id,
		resp.User.Email,
		resp.User.Name,
		resp.User.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.AuthPayload{
		Token: token,
		User:  protoToGraphQLUser(resp.User),
	}, nil
}

// Query resolver for current user

func (r *queryResolver) MeImpl(ctx context.Context) (*model.User, error) {
	// Get user from context (set by auth middleware)
	claims, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return nil, nil // Not authenticated, return nil (not an error)
	}

	// Fetch fresh user data from user-service
	req := &userpb.GetUserRequest{
		Id: claims.UserID,
	}

	resp, err := r.UserClient.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return protoToGraphQLUser(resp.User), nil
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

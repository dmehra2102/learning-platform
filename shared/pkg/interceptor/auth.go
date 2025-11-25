package interceptor

import (
	"context"
	"strings"

	"github.com/dmehra2102/learning-platform/shared/pkg/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
)

type AuthInterceptor struct {
	jwtManager    *jwt.Manager
	publicMethods map[string]bool
}

func NewAuthInterceptor(jwtManager *jwt.Manager) *AuthInterceptor {
	publicMethods := map[string]bool{
		"/user.UserService/Register":                 true,
		"/user.UserService/Login":                    true,
		"/course.CourseService/GetCourse":            true,
		"/course.CourseService/ListCourse":           true,
		"/review.ReviewService/ListCourseReviews":    true,
		"/review.ReviewService/GetCourseRatingStats": true,
	}

	return &AuthInterceptor{
		jwtManager:    jwtManager,
		publicMethods: publicMethods,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if i.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		newCtx, err := i.authorize(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if i.publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		newCtx, err := i.authorize(ss.Context())
		if err != nil {
			return err
		}

		wrappedStream := &wrappedServerStream{
			ctx:          newCtx,
			ServerStream: ss,
		}

		return handler(srv, wrappedStream)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	tokenString := values[0]
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := i.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
	ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

	return ctx, nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "user ID not found in context")
	}
	return userID, nil
}

func GetUserRole(ctx context.Context) (string, error) {
	role, ok := ctx.Value(UserRoleKey).(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "user role not found in context")
	}
	return role, nil
}

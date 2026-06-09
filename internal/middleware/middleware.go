package middleware

import (
	"context"

	"github.com/darktweek/cairn/internal/model"
)

type contextKey int

const (
	ctxKeyUser contextKey = iota
	ctxKeySession
)

func UserFromCtx(ctx context.Context) *model.User {
	u, _ := ctx.Value(ctxKeyUser).(*model.User)
	return u
}

func SessionFromCtx(ctx context.Context) *model.Session {
	s, _ := ctx.Value(ctxKeySession).(*model.Session)
	return s
}

func withUser(ctx context.Context, u *model.User) context.Context {
	return context.WithValue(ctx, ctxKeyUser, u)
}

func withSession(ctx context.Context, s *model.Session) context.Context {
	return context.WithValue(ctx, ctxKeySession, s)
}

package dctx

import (
	"context"
	"net/http"

	"github.com/SSHZ-ORG/dedicatus/dctx/protoconf"
	"github.com/SSHZ-ORG/dedicatus/dctx/protoconf/pb"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/appengine/v2"
)

type contextKey int

const (
	ctxKey contextKey = iota
)

type contextData struct {
	// The user in session. Might be nil.
	user *tgbotapi.User
	// UNKNOWN if no user in session.
	userType pb.AuthConfig_UserType

	conf *pb.Protoconf
}

func NewContext(req *http.Request) context.Context {
	ctx := appengine.NewContext(req)

	c, err := protoconf.GetConf(ctx)
	if err != nil {
		// Fail the request.
		panic("protoconf.GetConf: " + err.Error())
	}

	return context.WithValue(ctx, ctxKey, &contextData{
		conf: c,
	})
}

func fromContext(ctx context.Context) *contextData {
	if d, ok := ctx.Value(ctxKey).(*contextData); ok {
		return d
	}
	panic("attempting to call fromContext() without valid contextData")
}

func AttachUserInSession(ctx context.Context, user *tgbotapi.User) {
	d := fromContext(ctx)
	d.user = user
	d.userType = d.conf.GetAuthConfig().Users[int64(user.ID)]
}

func ProtoconfFromContext(ctx context.Context) *pb.Protoconf {
	return fromContext(ctx).conf
}

func UserFromContext(ctx context.Context) *tgbotapi.User {
	return fromContext(ctx).user
}

func UserTypeFromContext(ctx context.Context) pb.AuthConfig_UserType {
	return fromContext(ctx).userType
}

func IsAdmin(ctx context.Context) bool {
	d := fromContext(ctx)
	return d.userType == pb.AuthConfig_ADMIN
}

func IsContributor(ctx context.Context) bool {
	d := fromContext(ctx)
	return d.userType == pb.AuthConfig_ADMIN || d.userType == pb.AuthConfig_CONTRIBUTOR
}

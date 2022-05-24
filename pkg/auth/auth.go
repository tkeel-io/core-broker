package auth

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	tkeelTransutil "github.com/tkeel-io/kit/transport/http"
)

const (
	_XtKeelAuthUserHeader = "X-Tkeel-Auth"
	_AuthorizationHeader  = "Authorization"
)

var ErrNotFound = errors.New("authorization info not found")

type User struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Token    string `json:"token"`
	Auth     string `json:"auth"`
	TenantID string `json:"tenant_id"`
}

func GetUser(ctx context.Context) (User, error) {
	u := User{}
	headers := tkeelTransutil.HeaderFromContext(ctx)
	authHTTPHeader, ok := headers[_XtKeelAuthUserHeader]
	if !ok {
		return u, ErrNotFound
	}
	authInfo := strings.Join(authHTTPHeader, "")
	u.Auth = authInfo
	authStrBytes, err := base64.StdEncoding.DecodeString(authInfo)
	if err != nil {
		err = errors.Wrap(err, "decode auth header error")
		return u, err
	}
	q, err := url.ParseQuery(string(authStrBytes))
	if err != nil {
		err = errors.Wrap(err, "parse auth header error")
		return u, err
	}
	u.ID = q.Get("user")
	u.Role = q.Get("role")
	u.TenantID = q.Get("tenant")
	token, ok := headers[_AuthorizationHeader]
	if ok {
		u.Token = strings.Join(token, "")
	}
	return u, nil
}

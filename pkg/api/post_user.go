package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	apierrors "github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

type PostUserRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func (p *PostUserRequest) Validate() error {
	if p.Username == "" || p.Password == "" {
		return errors.New("'username' and 'password' must be provided in the request")
	}
	if p.Roles == nil || len(p.Roles) == 0 {
		return errors.New("You must assign at least one role to the user")
	}
	return nil
}

func (d *desktopAPI) CreateUser(w http.ResponseWriter, r *http.Request) {
	req := GetRequestObject(r).(*PostUserRequest)
	if err := req.Validate(); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	user := &types.User{
		Name:     req.Username,
		Password: req.Password,
		Roles:    make([]*types.Role, 0),
	}
	for _, role := range req.Roles {
		user.Roles = append(user.Roles, &types.Role{Name: role})
	}
	sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()
	if _, err := sess.GetUser(user.Name); err == nil {
		apiutil.ReturnAPIError(fmt.Errorf("A user with the name %s already exists", user.Name), w)
		return
	} else if !apierrors.IsUserNotFoundError(err) {
		apiutil.ReturnAPIError(err, w)
		return
	}
	if err := sess.CreateUser(user); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteOK(w)
}

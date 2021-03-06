package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

var elevateDenyReason = "The requested operation grants more privileges than the user has."

func denyUserElevatePerms(d *desktopAPI, reqUser *v1.VDIUser, r *http.Request) (allowed bool, reason string, err error) {

	// This is an ugly hack at the moment. This will be triggered if called from
	// allowSameUser while configuring MFA options. No need to check.
	if apiutil.GetGorillaPath(r) == "/api/users/{user}/mfa" {
		return true, "", nil
	}

	// Check that a POST /users will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.CreateUserRequest); ok {
		vdiRoles, err := d.vdiCluster.GetRoles(d.client)
		if err != nil {
			return false, "", err
		}
		for _, role := range reqObj.Roles {
			roleObj := getRoleByName(vdiRoles, role)
			if roleObj == nil {
				continue
			}
			for _, rule := range roleObj.GetRules() {
				if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
					return false, elevateDenyReason, nil
				}
			}
		}
		return true, "", nil
	}

	// Check that a PUT /users/{user} will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.UpdateUserRequest); ok {
		vdiRoles, err := d.vdiCluster.GetRoles(d.client)
		if err != nil {
			return false, "", err
		}
		for _, role := range reqObj.Roles {
			roleObj := getRoleByName(vdiRoles, role)
			if roleObj == nil {
				continue
			}
			for _, rule := range roleObj.GetRules() {
				if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
					return false, elevateDenyReason, nil
				}
			}
		}
		return true, "", nil
	}

	// Check that a POST /roles will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.CreateRoleRequest); ok {
		for _, rule := range reqObj.GetRules() {
			if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
				return false, elevateDenyReason, nil
			}
		}
		return true, "", nil
	}

	// Check that a PUT /roles/{role} will not grant permissions the user does not have.
	if reqObj, ok := apiutil.GetRequestObject(r).(*v1.UpdateRoleRequest); ok {
		for _, rule := range reqObj.GetRules() {
			if !reqUser.IncludesRule(rule, NewResourceGetter(d)) {
				return false, elevateDenyReason, nil
			}
		}
		return true, "", nil
	}

	apiLogger.Info("Method used privilege validator without adding request logic")
	return false, elevateDenyReason, nil
}

func getRoleByName(roles []v1alpha1.VDIRole, name string) *v1alpha1.VDIRole {
	for _, role := range roles {
		if role.GetName() == name {
			return &role
		}
	}
	return nil
}

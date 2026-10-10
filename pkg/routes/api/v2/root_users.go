package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

type rootUserBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type rootUserUpdateBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_provider"`
}

type rootUserStatusBody struct {
	Enabled bool `json:"enabled"`
}

type rootUserAdminBody struct {
	IsAdmin bool `json:"is_admin"`
}

type rootUserPasswordBody struct {
	Password string `json:"password"`
}

type rootUserListBody struct {
	Body []*rootUserView
}

type rootUserView struct {
	ID             int64       `json:"id"`
	Username       string      `json:"username"`
	Email          string      `json:"email,omitempty"`
	Status         user.Status `json:"status"`
	IsAdmin        bool        `json:"is_admin"`
	AvatarProvider string      `json:"avatar_provider"`
}

func newRootUserView(u *user.User) *rootUserView {
	return &rootUserView{
		ID: u.ID, Username: u.Username, Email: u.Email, Status: u.Status,
		IsAdmin: u.IsAdmin, AvatarProvider: u.AvatarProvider,
	}
}

func init() { AddRouteRegistrar(RegisterRootUserRoutes) }

func RegisterRootUserRoutes(api huma.API) {
	tags := []string{"root"}
	Register(api, huma.Operation{OperationID: "root-users-list", Summary: "List users", Description: "Lists all users for an authenticated instance administrator.", Method: http.MethodGet, Path: "/root/users", Tags: tags}, rootUsersList)
	Register(api, huma.Operation{OperationID: "root-users-create", Summary: "Create a user", Description: "Creates a local user account for an authenticated instance administrator.", Method: http.MethodPost, Path: "/root/users", Tags: tags}, rootUsersCreate)
	Register(api, huma.Operation{OperationID: "root-users-update", Summary: "Update a user", Description: "Updates a user's username, email, or avatar provider.", Method: http.MethodPut, Path: "/root/users/{id}", Tags: tags}, rootUsersUpdate)
	Register(api, huma.Operation{OperationID: "root-users-password", Summary: "Reset a password", Description: "Sets a user's password directly.", Method: http.MethodPost, Path: "/root/users/{id}/password", Tags: tags}, rootUsersPassword)
	Register(api, huma.Operation{OperationID: "root-users-password-email", Summary: "Send a password reset", Description: "Sends the user a password reset email.", Method: http.MethodPost, Path: "/root/users/{id}/password-reset", Tags: tags}, rootUsersPasswordEmail)
	Register(api, huma.Operation{OperationID: "root-users-status", Summary: "Change user status", Description: "Enables or disables a user account.", Method: http.MethodPost, Path: "/root/users/{id}/status", Tags: tags}, rootUsersStatus)
	Register(api, huma.Operation{OperationID: "root-users-admin", Summary: "Change admin status", Description: "Promotes or demotes a user without the licensed feature gate.", Method: http.MethodPost, Path: "/root/users/{id}/admin", Tags: tags}, rootUsersAdmin)
	Register(api, huma.Operation{OperationID: "root-users-delete", Summary: "Delete a user", Description: "Deletes immediately or schedules deletion for a user.", Method: http.MethodDelete, Path: "/root/users/{id}", Tags: tags}, rootUsersDelete)
}

func rootAdmin(ctx context.Context) (*user.User, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if !u.IsAdmin {
		return nil, huma.Error403Forbidden("instance administrator required")
	}
	return u, nil
}

func rootUsersList(ctx context.Context, _ *struct{}) (*rootUserListBody, error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	users, err := user.ListAllUsers(s)
	if err != nil {
		return nil, translateDomainError(err)
	}
	views := make([]*rootUserView, 0, len(users))
	for _, u := range users {
		views = append(views, newRootUserView(u))
	}
	return &rootUserListBody{Body: views}, nil
}

func rootUsersCreate(ctx context.Context, in *struct{ Body rootUserBody }) (*singleBody[rootUserView], error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)
	u, err := user.CreateUser(s, &user.User{Username: in.Body.Username, Email: in.Body.Email, Password: in.Body.Password})
	if err == nil {
		err = models.CreateNewProjectForUser(s, u)
	}
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &singleBody[rootUserView]{Body: newRootUserView(u)}, nil
}

func rootUserByID(s *xorm.Session, id int64) (*user.User, error) { return user.GetUserByID(s, id) }

func rootUsersUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body rootUserUpdateBody
}) (*singleBody[rootUserView], error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	u, err := rootUserByID(s, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	u.Username, u.Email, u.AvatarProvider = in.Body.Username, in.Body.Email, in.Body.Avatar
	u, err = user.UpdateUser(s, u, false)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[rootUserView]{Body: newRootUserView(u)}, nil
}

func rootUsersPassword(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body rootUserPasswordBody
}) (*singleBody[rootUserView], error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	u, err := rootUserByID(s, in.ID)
	if err == nil {
		err = user.UpdateUserPassword(s, u, in.Body.Password)
	}
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[rootUserView]{Body: newRootUserView(u)}, nil
}

func rootUsersPasswordEmail(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	u, err := rootUserByID(s, in.ID)
	if err == nil {
		err = user.RequestUserPasswordResetToken(s, u)
	}
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &emptyBody{}, nil
}

func rootUsersStatus(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body rootUserStatusBody
}) (*singleBody[rootUserView], error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	u, err := rootUserByID(s, in.ID)
	if err == nil {
		status := user.StatusDisabled
		if in.Body.Enabled {
			status = user.StatusActive
		}
		err = models.ChangeUserStatus(s, nil, u, status)
	}
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &singleBody[rootUserView]{Body: newRootUserView(u)}, nil
}

func rootUsersAdmin(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body rootUserAdminBody
}) (*singleBody[rootUserView], error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	u, err := rootUserByID(s, in.ID)
	if err == nil && !in.Body.IsAdmin {
		err = user.GuardLastAdmin(s, u)
	}
	if err == nil {
		u.IsAdmin = in.Body.IsAdmin
		_, err = s.ID(u.ID).Cols("is_admin").Update(u)
	}
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[rootUserView]{Body: newRootUserView(u)}, nil
}

func rootUsersDelete(ctx context.Context, in *struct {
	ID   int64  `path:"id"`
	Mode string `query:"mode"`
}) (*emptyBody, error) {
	if _, err := rootAdmin(ctx); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	u, err := rootUserByID(s, in.ID)
	if err == nil {
		err = user.GuardLastAdmin(s, u)
	}
	if err == nil && in.Mode == "scheduled" {
		err = user.RequestDeletion(s, u)
	} else if err == nil {
		err = models.DeleteUser(s, u)
	}
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &emptyBody{}, nil
}

package main

import (
	"html/template"
	"sort"
	"strings"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp"
	"github.com/andreyvit/mvp/forms"

	m "github.com/andreyvit/buddyd/model"
)

type (
	UserStatusGroupVM struct {
		Title string
		Users []*UserWithMembershipVM
	}

	UserWithMembershipVM struct {
		*m.User
		Membership *m.UserMembership
	}
)

func (app *App) listAdminUsers(rc *RC, in *struct{}) (*mvp.ViewData, error) {
	accountID := rc.AccountID()
	var admins, assistants, regulars, invited, banned, rejected []*UserWithMembershipVM
	for c := edb.ExactIndexScan[m.User](rc, UsersByAccount, accountID); c.Next(); {
		u := c.Row()
		if memb := u.Membership(accountID); memb != nil && memb.Status.IsKnown() {
			um := &UserWithMembershipVM{
				User:       u,
				Membership: memb,
			}
			if memb.Status == m.UserStatusBanned {
				banned = append(banned, um)
			} else if memb.Status == m.UserStatusSelfRejected {
				rejected = append(rejected, um)
			} else if memb.Role == m.UserAccountRoleOwner {
				admins = append(admins, um)
			} else if memb.Role == m.UserAccountRoleAdmin {
				admins = append(admins, um)
			} else if memb.Role == m.UserAccountRoleAssistant {
				assistants = append(assistants, um)
			} else if memb.Status == m.UserStatusInvited {
				invited = append(invited, um)
			} else if memb.Role == m.UserAccountRoleConsumer && memb.Status == m.UserStatusActive {
				regulars = append(regulars, um)
			}
		}
	}

	var groups []*UserStatusGroupVM
	if len(admins) > 0 {
		groups = append(groups, &UserStatusGroupVM{
			Title: "Admins",
			Users: admins,
		})
	}
	if len(assistants) > 0 {
		groups = append(groups, &UserStatusGroupVM{
			Title: "Coaches",
			Users: assistants,
		})
	}
	if len(regulars) > 0 {
		groups = append(groups, &UserStatusGroupVM{
			Title: "Regular Users",
			Users: regulars,
		})
	}
	if len(invited) > 0 {
		groups = append(groups, &UserStatusGroupVM{
			Title: "Invited Users",
			Users: invited,
		})
	}
	if len(banned) > 0 {
		groups = append(groups, &UserStatusGroupVM{
			Title: "Banned",
			Users: banned,
		})
	}

	return &mvp.ViewData{
		View:         "admin/users",
		Title:        "Users",
		SemanticPath: "admin/users",
		Data: struct {
			Groups []*UserStatusGroupVM
		}{
			Groups: groups,
		},
	}, nil
}

func (app *App) handleAdminWhitelist(rc *mvp.RC, in *struct {
	IsSaving bool `json:"-" form:",issave"`
}) (any, error) {
	accountID := fullRC.From(rc).AccountID()
	all := make(map[string]*m.User)
	whitelisted := make(map[string]*m.User)
	var whitelist []string
	for c := edb.ExactIndexScan[m.User](rc, UsersByAccount, accountID); c.Next(); {
		u := c.Row()
		if in.IsSaving {
			all[u.EmailNorm] = u
		}
		if memb := u.Membership(accountID); memb != nil && memb.Status.ActiveOrInvited() && memb.Source == m.UserSourceWhitelist {
			whitelist = append(whitelist, u.Email)
			whitelisted[u.EmailNorm] = u
		}
	}

	sort.Strings(whitelist)
	whitelistStr := strings.Join(whitelist, "\n")

	form := &forms.Form{
		Multipart: true,
		Group: forms.Group{
			Styles: []*forms.Style{
				adminFormStyle,
				horizontalFormStyle,
			},
			Children: []forms.Child{
				&forms.Group{
					Children: []forms.Child{
						&forms.Item{
							Name:  "whitelist",
							Label: "Emails allowed to sign up",
							Child: &forms.InputText{
								Template: "control-textarea",
								TagOpts: forms.TagOpts{
									Class: "",
									Attrs: map[string]any{"rows": 15},
								},
								Binding:     forms.Var(&whitelistStr),
								Placeholder: "",
							},
						},
					},
				},

				saveFormButtonBar(),
			},
		},
	}

	if in.IsSaving && form.ProcessRequest(rc.Request.Request) {
		for _, email := range strings.Fields(whitelistStr) {
			canon := mvp.CanonicalEmail(email)
			u := all[canon]
			var modified bool
			if u == nil {
				u = &m.User{
					ID:        app.NewID(),
					Role:      m.UserSystemRoleRegular,
					Email:     email,
					EmailNorm: canon,
				}
				modified = true
			}
			memb := u.Membership(accountID)
			if memb == nil {
				memb = &m.UserMembership{
					CreationTime: rc.Now,
					AccountID:    accountID,
					Role:         m.UserAccountRoleConsumer,
				}
				u.Memberships = append(u.Memberships, memb)
				modified = true
			}
			if memb.Source != m.UserSourceWhitelist {
				memb.Source = m.UserSourceWhitelist
				modified = true
			}
			if memb.Status.Invitable() {
				memb.Status = m.UserStatusActive // TODO: invitation flow?
				modified = true
			}
			if modified {
				edb.Put(rc, u)
			}
			delete(whitelisted, u.EmailNorm)
		}
		for _, u := range whitelisted {
			memb := u.Membership(accountID)
			if memb != nil && memb.Status.ActiveOrInvited() {
				memb.Status = m.UserStatusInactive
				edb.Put(rc, u)
			}
		}
		return rc.Redirect("admin.whitelist"), nil
	}

	return &mvp.ViewData{
		View:         "form",
		Title:        "Whitelist",
		SemanticPath: "admin/whitelist",
		Data: struct {
			Form template.HTML
		}{
			Form: app.RenderForm(rc, form),
		},
	}, nil
}

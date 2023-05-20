package main

import (
	"html/template"
	"sort"
	"strings"

	m "github.com/andreyvit/buddyd/model"
	"github.com/andreyvit/buddyd/mvp"
	"github.com/andreyvit/buddyd/mvp/forms"
	"github.com/andreyvit/edb"
)

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
			WrapperTag: forms.TagOpts{
				Class: "my-16",
			},
			Children: []forms.Child{
				&forms.Header{
					Text: "Whitelist",
				},
				&forms.Group{
					Children: []forms.Child{
						&forms.Item{
							Name:  "whitelist",
							Label: "Whitelisted Emails",
							Child: &forms.InputText{
								Template: "control-textarea",
								TagOpts: forms.TagOpts{
									Class: "",
									Attrs: map[string]any{"rows": 3},
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
		View:         "admin/whitelist",
		Title:        "Whitelist",
		SemanticPath: "admin/whitelist",
		Data: struct {
			Form template.HTML
		}{
			Form: app.RenderForm(rc, form),
		},
	}, nil
}

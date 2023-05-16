package m

type Permission int

const (
	PermissionNone = Permission(iota)

	PermissionAccessSuperadminArea
	PermissionManageSuperadmins

	PermissionAccessAdminArea
	PermissionManageAccount
)

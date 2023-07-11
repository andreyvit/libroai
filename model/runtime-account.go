package m

type (
	RuntimeAccount struct {
		*Account

		UsersByID map[UserID]*User
	}

	RuntimeUser struct {
		*User
	}
)

func (racc *RuntimeAccount) UserByID(id UserID) *User {
	return racc.UsersByID[id]
}

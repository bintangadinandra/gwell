package user

// User ...
type User struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// LoginRequest ...
type LoginRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

// Login ...
func Login(request *LoginRequest) bool {
	predefinedUsers := []*User{
		&User{
			UserName: "bintang",
			Password: "bintang123",
		},
		&User{
			UserName: "juan",
			Password: "juan123",
		},
		&User{
			UserName: "sebas",
			Password: "sebas123",
		},
		&User{
			UserName: "sarip",
			Password: "sarip123",
		},
		&User{
			UserName: "bitcan01",
			Password: "b1Tc4NzeR0on3",
		},
	}
	for _, predefinedUser := range predefinedUsers {
		if predefinedUser.UserName == request.UserName && predefinedUser.Password == request.Password {
			return true
		}
	}
	return false
}

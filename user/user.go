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
		&User{
			UserName: "bitcan02",
			Password: "8esu6ce4ah",
		},
		&User{
			UserName: "bitcan03",
			Password: "yh5af6jy5g",
		},
		&User{
			UserName: "bitcan04",
			Password: "g3qdywpkcy",
		},
		&User{
			UserName: "bitcan05",
			Password: "x7du9vpuq7",
		},
		&User{
			UserName: "bitcan06",
			Password: "wt73tcnt6h",
		},
		&User{
			UserName: "bitcan07",
			Password: "46fusrzemg",
		},
		&User{
			UserName: "bitcan08",
			Password: "c79k4kbrya",
		},
		&User{
			UserName: "bitcan09",
			Password: "stqpduc99v",
		},
		&User{
			UserName: "bitcan10",
			Password: "72zf6xnutg",
		},
	}
	for _, predefinedUser := range predefinedUsers {
		if predefinedUser.UserName == request.UserName && predefinedUser.Password == request.Password {
			return true
		}
	}
	return false
}

package db

func CreateUser(u *User) error {
	_, err := db.Exec(
		"INSERT INTO users(pubkey) VALUES ($1) ON CONFLICT(pubkey) DO UPDATE SET last_seen = CURRENT_TIMESTAMP",
		u.Pubkey)
	return err
}

func UpdateUser(u *User) error {
	_, err := db.Exec("UPDATE users SET last_seen = $1 WHERE pubkey = $2", u.LastSeen, u.Pubkey)
	return err
}

package db

func (db *DB) CreateUser(u *User) error {
	_, err := db.Exec(
		"INSERT INTO users(pubkey) VALUES ($1) ON CONFLICT(pubkey) DO UPDATE SET last_seen = CURRENT_TIMESTAMP",
		u.Pubkey)
	return err
}

func (db *DB) FetchUser(pubkey string, u *User) error {
	return db.QueryRow("SELECT pubkey, last_seen FROM users WHERE pubkey = $1", pubkey).Scan(&u.Pubkey, &u.LastSeen)
}

func (db *DB) UpdateUser(u *User) error {
	_, err := db.Exec("UPDATE users SET last_seen = $1 WHERE pubkey = $2", u.LastSeen, u.Pubkey)
	return err
}

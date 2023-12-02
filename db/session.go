package db

func (db *DB) CreateSession(s *Session) error {
	_, err := db.Exec("INSERT INTO sessions(pubkey, session_id) VALUES($1, $2)", s.Pubkey, s.SessionId)
	return err
}

func (db *DB) FetchSession(s *Session) error {
	query := "SELECT u.pubkey, u.msats FROM sessions s LEFT JOIN users u ON u.pubkey = s.pubkey WHERE session_id = $1"
	err := db.QueryRow(query, s.SessionId).Scan(&s.Pubkey, &s.Msats)
	return err
}

func (db *DB) DeleteSession(s *Session) error {
	_, err := db.Exec("DELETE FROM sessions where session_id = $1", s.SessionId)
	return err
}

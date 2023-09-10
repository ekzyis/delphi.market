package db

func CreateSession(s *Session) error {
	_, err := db.Exec("INSERT INTO sessions(pubkey, session_id) VALUES($1, $2)", s.Pubkey, s.SessionId)
	return err
}

func FetchSession(s *Session) error {
	err := db.QueryRow("SELECT pubkey FROM sessions WHERE session_id = $1", s.SessionId).Scan(&s.Pubkey)
	return err
}

func DeleteSession(s *Session) error {
	_, err := db.Exec("DELETE FROM sessions where session_id = $1", s.SessionId)
	return err
}

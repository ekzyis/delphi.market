package db

func CreateLNAuth(lnAuth *LNAuth) error {
	err := db.QueryRow(
		"INSERT INTO lnauth(k1, lnurl) VALUES($1, $2) RETURNING session_id",
		lnAuth.K1, lnAuth.LNURL).Scan(&lnAuth.SessionId)
	return err
}

func FetchSessionId(k1 string, sessionId *string) error {
	err := db.QueryRow("SELECT session_id FROM lnauth WHERE k1 = $1", k1).Scan(sessionId)
	return err
}

func DeleteLNAuth(lnAuth *LNAuth) error {
	_, err := db.Exec("DELETE FROM lnauth WHERE k1 = $1", lnAuth.K1)
	return err
}

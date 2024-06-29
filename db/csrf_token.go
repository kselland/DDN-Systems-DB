package db

type CsrfToken struct {
	Token string
	User_Id int
}

func InsertCsrfToken(d CsrfToken) error {
	_, err := db.Exec("INSERT INTO csrf_tokens (token, user_id) VALUES ($1, $2)", d.Token, d.User_Id)
    return err
}

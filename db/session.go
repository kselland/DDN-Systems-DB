package db

import (
	"errors"
	"strconv"
)

type SessionInsertion struct {
    SessionKeyDigest []byte
    CsrfToken string
    UserId int
}
func InsertSession(si SessionInsertion) (*int, error) {
	query, err := db.Query(`
		INSERT INTO sessions (
			session_key_digest,
			csrf_token,
			user_id
		) VALUES (
			$1,
			$2,
			$3
		) RETURNING
			id`,
		si.SessionKeyDigest,
		si.CsrfToken,
		si.UserId,
	)
	if err != nil {
		return nil, err
	}
	res, err := getFirst[struct{ Id string }](query)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("SQL isn't working properly")
	}

    sessionId, err := strconv.Atoi(res.Id)
	if err != nil {
		return nil, err
	}

    return &sessionId, nil
}

type flatSession struct {
	Id                   int
	User_Id              int
	Session_Key_Digest   []byte
	Csrf_Token           string
	User_Role            UserRole
	User_Email           string
	User_Password_Digest []byte
	User_Name            string
}

type Session struct {
	Id                 int
	User               User
	Session_Key_Digest []byte
	Csrf_Token         string
}

func flatSessionToSession(flatSession *flatSession) Session {
	return Session{
		Id: flatSession.Id,
		User: User{
			Id:              flatSession.User_Id,
			Email:           flatSession.User_Email,
			Password_digest: flatSession.User_Password_Digest,
			Role:            flatSession.User_Role,
			Name:            flatSession.User_Name,
		},
		Session_Key_Digest: flatSession.Session_Key_Digest,
		Csrf_Token:         flatSession.Csrf_Token,
	}
}

func GetSession(id int) (*Session, error) {
	query, err := db.Query(`
		SELECT
			s.id,
			s.user_id,
			s.session_key_digest,
			s.csrf_token,
			u.role user_role,
			u.email user_email,
			u.password_digest user_password_digest,
			u.name user_name
		FROM
			sessions s
		LEFT JOIN
			users u
		ON
			s.user_id = u.id
		WHERE
			s.id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	flatSession, err := getFirst[flatSession](query)
	if err != nil {
		return nil, err
	}
	if flatSession == nil {
		return nil, errors.New("Couldn't find session")
	}

    session := flatSessionToSession(flatSession)

	return &session, nil
}





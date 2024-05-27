package session

import (
	"bytes"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"errors"
	"log"
	"net/http"
)

type Session struct {
	Id                 int
	User_Id            int
	Session_Key_Digest []byte
	Csrf_Token         string
}

func AuthenticateSession(r *http.Request) (*Session, error) {
	sessionIdCookie, err := r.Cookie("SESSION_ID")
	if err != nil {
		return nil, errors.New("No session id")
	}

	sessionKeyCookie, err := r.Cookie("SESSION_KEY")
	if err != nil {
		return nil, errors.New("No session Key")
	}

	query, err := db.Db.Query("SELECT id, user_id, session_key_digest, csrf_token FROM sessions WHERE id = $1", sessionIdCookie.Value)
	if err != nil {
		return nil, err
	}

	session, err := db.GetFirst[Session](query)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("No session with that id")
	}

	digestOfSessionKeyCookie, err := lib.GetDigest(sessionKeyCookie.Value)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(digestOfSessionKeyCookie, session.Session_Key_Digest) {
		return nil, errors.New("Digest didn't match")
	}

	return session, nil
}

func CreateSession(w http.ResponseWriter, userId int) error {
	sessionKey, err := lib.GenerateToken()
	if err != nil {
		return err
	}
    csrfToken, err := lib.GenerateToken()
    if err != nil {
        return err
    }
    sessionKeyDigest, err := lib.GetDigest(sessionKey)
    if err != nil {
        return err
    }

    query, err := db.Db.Query("INSERT INTO sessions (session_key_digest, csrf_token, user_id) VALUES ($1, $2, $3) RETURNING id", sessionKeyDigest, csrfToken, userId)
    if err != nil {
        return err
    }

    res, err := db.GetFirst[struct {Id string}](query)
	if err != nil {
		return err
	}
    if res == nil {
        return errors.New("SQL isn't working properly")
    }

    sessionId := res.Id

	log.Println("Made it here as well")
	http.SetCookie(w, &http.Cookie{
		Name:  "SESSION_ID",
		Value: sessionId,
		Path:  "/",
		MaxAge: 2000,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "SESSION_KEY",
		Value: sessionKey,
		Path:  "/",
		MaxAge: 2000,
	})

    return nil
}

func EndSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
        Name: "SESSION_ID",
        Path: "/",
        MaxAge: -1,
    })
    http.SetCookie(w, &http.Cookie{
        Name: "SESSION_KEY",
        Path: "/",
        MaxAge: -1,
    })
}
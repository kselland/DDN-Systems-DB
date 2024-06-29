package session

import (
	"bytes"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"net/http"
	"strconv"
)

/*
If there is an error while trying to authenticate (like fail to hit db or smnth)
then `error` will be populated. If authentication fails then the `*Session` will be nill
there will also be no error however
*/
func AuthenticateSession(r *http.Request) (*db.Session, error) {
	sessionIdCookie, err := r.Cookie("SESSION_ID")
	if err != nil {
		return nil, nil
	}

	sessionKeyCookie, err := r.Cookie("SESSION_KEY")
	if err != nil {
		return nil, nil
	}

	sessionId, err := strconv.Atoi(sessionIdCookie.Value)
	if err != nil {
		return nil, nil
	}

	session, err := db.GetSession(sessionId)
	if err != nil {
		return nil, err
	}

	digestOfSessionKeyCookie, err := lib.GetDigest(sessionKeyCookie.Value)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(digestOfSessionKeyCookie, session.Session_Key_Digest) {
		return nil, nil
	}

	return session, nil
}

func CreateSession(r *http.Request, w http.ResponseWriter, userId int) error {
	if session, err := AuthenticateSession(r); err == nil && session != nil && session.User.Id == userId {
		sessionKeyCookie, err := r.Cookie("SESSION_KEY")
		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:   "SESSION_ID",
				Value:  strconv.Itoa(session.Id),
				Path:   "/",
				MaxAge: 2000,
			})
			http.SetCookie(w, &http.Cookie{
				Name:   "SESSION_KEY",
				Value:  sessionKeyCookie.Value,
				Path:   "/",
				MaxAge: 2000,
			})
			return nil
		}
	}

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

	sessionId, err := db.InsertSession(db.SessionInsertion {
		SessionKeyDigest: sessionKeyDigest,
		CsrfToken: csrfToken,
		UserId: userId,
	})
	if err != nil {
		return err
	}


	http.SetCookie(w, &http.Cookie{
		Name:   "SESSION_ID",
		Value:  strconv.Itoa(*sessionId),
		Path:   "/",
		MaxAge: 2000,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "SESSION_KEY",
		Value:  sessionKey,
		Path:   "/",
		MaxAge: 2000,
	})

	return nil
}

func EndSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "SESSION_ID",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "SESSION_KEY",
		Path:   "/",
		MaxAge: -1,
	})
}

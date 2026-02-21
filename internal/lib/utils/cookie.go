package utils

import (
	"net/http"
)

func GetCookieValue(r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func SetCookie(w http.ResponseWriter, name, value string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: httpOnly,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
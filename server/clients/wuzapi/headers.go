package wuzapi

import "net/http"

func SetUserToken(r *http.Request, token string) {
	r.Header.Set("token", token)          // o certo
	r.Header.Set("Authorization", token)  // fallback de build esquisita
}
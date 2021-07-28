package http

import (
	"fmt"
	"net/http"
)

func (a *App) myMiddleware(handler http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("middleware: before handler call")
		handler.ServeHTTP(w, r)
		fmt.Println("middleware: after handler call")
	}

	return http.HandlerFunc(f)
}

func (a *App) addUserToSession(handler http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		exists := a.Session.Exists(r, "userID")
		if !exists {
			handler.ServeHTTP(w, r)
			return
		}

		uid := a.Session.GetInt(r, "userID")
		if uid == 0 {
			// todo:: handle error
			fmt.Println("error: uid is zero")
			return
		}

		fmt.Println("user id:", uid)

		u, err := a.UserService.User(uid)
		if err != nil {
			// todo:: handle error
			fmt.Println(err)
			return
		}
		fmt.Println("user id:", u)
		handler.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
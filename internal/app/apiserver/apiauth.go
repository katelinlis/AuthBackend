package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/katelinlis/AuthBackend/internal/app/model"
)

func (s *server) ConfigurAuthRouter() {

	router := s.router.PathPrefix("/api/auth").Subrouter()
	router.HandleFunc("/create", s.HandleCreateUser()).Methods("POST") // Регистрация
	router.HandleFunc("/auth", s.HandleAuth()).Methods("POST")         // Авторизация
	router.HandleFunc("/refresh", s.HandleRefresh()).Methods("POST")   // Обновление JWT и получение нового Refresh токена
	router.HandleFunc("/sessions", s.HandleSessions()).Methods("GET")  // список сессий
}

// Register user request
type Register struct {
	Password  string `json:"text"`
	Login     string `json:"username"`
	Recaptcha string `json:"captcha"`
}

//AuthJWTRT object
type AuthJWTRT struct {
	Jwt          string `json:"jwt"`
	RefreshToken string `json:"refresh_token"`
}

func (s *server) HandleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var createPost Register
		json.NewDecoder(request.Body).Decode(&createPost)

		var auth model.Auth
		auth.Login = createPost.Login
		auth.Password = createPost.Password
		err := s.store.Auth().Create(&auth)

		if err != nil {
			s.respond(w, request, http.StatusUnprocessableEntity, err)
			return
		}

		jwt, err := s.jwtKeys.Create(auth.ID)

		RT, err := s.jwtKeys.CreateRT(auth.ID)

		s.respond(w, request, http.StatusOK, AuthJWTRT{
			Jwt:          jwt,
			RefreshToken: RT,
		})
	}
}

// Auth user request
type Auth struct {
	Password  string `json:"password"`
	Login     string `json:"username"`
	Recaptcha string `json:"captcha"`
}

func (s *server) HandleAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var createPost Auth
		json.NewDecoder(request.Body).Decode(&createPost)

		var auth model.Auth
		auth.Login = createPost.Login
		auth.Password = createPost.Password
		err := s.store.Auth().GetUserByUsername(&auth)

		compare := auth.ComparePassword()
		if !compare {
			s.respond(w, request, http.StatusUnprocessableEntity, "incorrect password")
			return
		}

		if err != nil {
			s.respond(w, request, http.StatusUnprocessableEntity, err)
			return
		}

		jwt, err := s.jwtKeys.Create(auth.ID)
		RT, err := s.jwtKeys.CreateRT(auth.ID)

		s.respond(w, request, http.StatusOK, AuthJWTRT{
			Jwt:          jwt,
			RefreshToken: RT,
		})
	}
}

// Refresh user request
type Refresh struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *server) HandleRefresh() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		userid, err := s.GetDataFromToken(w, request)
		if err != nil {
			fmt.Println(err)
		}
		var createPost Refresh
		json.NewDecoder(request.Body).Decode(&createPost)

		var auth model.Auth
		auth.Login = createPost.RefreshToken
		err = s.store.Auth().Create(&auth)

		if err != nil {
			s.respond(w, request, http.StatusUnprocessableEntity, err)
			return
		}

		s.redis.Del("wallget/" + string(rune(int(userid))))
		s.respond(w, request, http.StatusOK, auth)
	}
}

func (s *server) HandleSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
	}
}

package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/katelinlis/AuthBackend/internal/app/store"
	"github.com/sirupsen/logrus"
)

type server struct {
	router  *mux.Router
	logger  *logrus.Logger
	store   store.Store
	redis   *redis.Client
	jwtKeys JwtKeys
}

const (
	ctxKeyUser ctxKey = iota
)

type ctxKey int8

var (
	errIncorrectEmailOrPassword = errors.New("incorect email or password")
	jwtsignkey                  string
)

func newServer(store store.Store, config *Config) *server {
	s := &server{
		jwtKeys: getJwtKeys("./"),
		router:  mux.NewRouter(),
		logger:  logrus.New(),
		redis: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		store: store,
	}
	s.configureRouter()

	jwtsignkey = config.JwtSignKey

	return s
}

func (s *server) GetDataFromToken(w http.ResponseWriter, r *http.Request) (float64, error) {
	var token string
	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
	}

	if token == "" {
		return 0, errors.New("Token is missing")
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			msg := fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			return 0, msg
		}
		return []byte(jwtsignkey), nil
	})

	if err != nil {
		s.error(w, r, http.StatusUnauthorized, errors.New("Error parsing token"))
		return 0, err
	}
	if parsedToken != nil && parsedToken.Valid {
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
			userid := claims["userid"].(float64)
			return userid, nil
		}
	}
	return 0, nil

}

//Get Limit and Offset from request
func (s *server) URLLimitOffset(request *http.Request) (int, int) {
	var offset int
	var limit = 20
	urlParams := request.URL.Query()
	if len(urlParams["offset"]) > 0 {
		offsetInt, err := strconv.Atoi(urlParams["offset"][0])
		if err == nil {
			offset = offsetInt
		}
	}
	if len(urlParams["limit"]) > 0 {
		limitInt, err := strconv.Atoi(urlParams["limit"][0])
		if err == nil {
			limit = limitInt
		}
	}
	return offset, limit
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {

	s.router.Use(s.loggingMiddleware)

	s.ConfigurAuthRouter()
}

func (s *server) emptyresponse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Request " + r.RequestURI + " from " + r.RemoteAddr)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)

	r.Body.Close()

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

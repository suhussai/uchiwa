package authentication

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/sensu/uchiwa/uchiwa/audit"
	"github.com/sensu/uchiwa/uchiwa/helpers"
	"github.com/sensu/uchiwa/uchiwa/structs"
	log "github.com/Sirupsen/logrus"
)

// New function initalizes and returns a Config struct
func New(auth structs.Auth) Config {
	a := Config{
		Auth: auth,
	}
	return a
}

// publicHandler does not enforce authentication
func publicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// restrictedHandler enforce authentication by validating the JWT
// or the access token provided in the configuration
func restrictedHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the JWT
		token, err := verifyJWT(r)
		if err != nil {
			log.Debug("No JWT token provided")
		}

		// Verify the access token if no JWT was provided
		if err != nil {
			token, err = verifyAccessToken(r)
		}

		// If no JWT or access token found
		if err != nil {
			log.Debug("No access token provided")
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		setJWTInContext(r, token)
		next.ServeHTTP(w, r)
		context.Clear(r)
		return
	})
}

// Authenticate calls the proper handler based on whether authentication is enabled or not
func (a *Config) Authenticate(next http.Handler) http.Handler {
	if a.DriverName == "none" {
		return publicHandler(next)
	}
	return restrictedHandler(next)
}

// Login authenticates a user against the authentication driver
func (a *Config) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			decoder := json.NewDecoder(r.Body)
			var data interface{}
			err := decoder.Decode(&data)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Warn("Could not decode the body.")
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			m, ok := data.(map[string]interface{})
			if !ok {
				log.WithFields(log.Fields{
					"error": err,
				}).Warn("Could not assert the body.")
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			u := m["user"].(string)
			p := m["pass"].(string)
			if u == "" || p == "" {
				log.Info("Authentication failed: user and password must not be empty")
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			user, err := a.login(u, p)
			if err != nil {
				log.Info(err)

				// Output to audit log
				log := structs.AuditLog{Action: "loginfailure", Level: "default", Output: err.Error()}
				log.RemoteAddr = helpers.GetIP(r)
				audit.Log(log)

				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			// Obfuscate user attributes
			user.Password = ""

			j, err := json.Marshal(user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(j)
			return
		}

		http.Redirect(w, r, "/#/login", http.StatusFound)
		return
	})
}

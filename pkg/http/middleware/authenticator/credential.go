package authenticator

import (
	"net/http"

	"github.com/vardius/go-api-boilerplate/pkg/errors"
	"github.com/vardius/go-api-boilerplate/pkg/http/response"
	"github.com/vardius/go-api-boilerplate/pkg/identity"
)

// CredentialsAuthFunc returns Identity from username and password combination
type CredentialsAuthFunc func(username, password string) (identity.Identity, error)

// CredentialsAuthenticator authorize by the username and password
// and adds Identity to request's Context
type CredentialsAuthenticator interface {
	// FromBasicAuth authorize by the username and password provided in the request's
	// Authorization header, if the request uses HTTP Basic Authentication.
	FromBasicAuth(next http.Handler) http.Handler
}

type credentialsAuth struct {
	afn CredentialsAuthFunc
}

func (a *credentialsAuth) FromBasicAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		username, password, hasAuth := r.BasicAuth()
		if hasAuth {
			i, err := a.afn(username, password)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Basic`)

				appErr := errors.New(errors.UNAUTHORIZED, http.StatusText(http.StatusUnauthorized))
				w.WriteHeader(errors.HTTPStatusCode(appErr))

				if err := response.JSON(r.Context(), w, appErr); err != nil {
					panic(err)
				}
				return
			}

			next.ServeHTTP(w, r.WithContext(identity.ContextWithIdentity(r.Context(), i)))
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// NewCredentials returns new credentials authenticator
func NewCredentials(afn CredentialsAuthFunc) CredentialsAuthenticator {
	return &credentialsAuth{
		afn: afn,
	}
}

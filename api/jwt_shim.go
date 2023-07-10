package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// wraps jwtauth.JWTAuth with helper functions to ease usage of JWTs for API authentication
type JWT struct {
	tokenAuth           *jwtauth.JWTAuth
	identifierField     string
	validityDurationSec int64
}

// Initializes a new JWT object, with the signature algorith m HS256 encrypted with the given `password`.
// If you want to add supplemental information to the JWT set `identifierField` to some string value.
// The tokens will be valid for a number of seconds equal to `tokenValidityDurationSec`
func NewJWT(password string, identifierField string, tokenValidityDurationSec int64) *JWT {
	return NewJWTWithSignature(password, identifierField, "HS256", tokenValidityDurationSec)
}

// Like NewJWT but allows control of the signature algorithm.
func NewJWTWithSignature(password string, identifierField string, signature string, tokenValidityDurationSec int64) *JWT {
	return &JWT{
		identifierField:     identifierField,
		tokenAuth:           jwtauth.New(signature, []byte(password), nil),
		validityDurationSec: tokenValidityDurationSec,
	}
}

// Issues a new jwt adding the given identifier and extra fields to the claims.
func (jwt *JWT) Encode(identifier string, extraFields map[string]interface{}) (string, error) {
	if extraFields == nil {
		extraFields = make(map[string]interface{}, 3)
	}
	expiresAt := time.Duration(time.Now().Unix() + (int64(time.Second) * jwt.validityDurationSec))
	if identifier != "" {
		extraFields[jwt.identifierField] = identifier
	}
	jwtauth.SetIssuedNow(extraFields)
	jwtauth.SetExpiryIn(extraFields, expiresAt)

	_, encoded, err := jwt.tokenAuth.Encode(extraFields)
	if err != nil {
		return "", fmt.Errorf("failed to encode jwt: %s", err)
	}
	return encoded, nil
}

// Parses an encoded jwt token it's a jwt.Token type.
func (jwt *JWT) Decode(ctx context.Context, token string) (jwt.Token, error) {
	return jwt.tokenAuth.Decode(token)
}

// Used to perform validation of the jwt token, and identifier fields if required.
func (jt *JWT) CheckToken(token jwt.Token) error {
	if token == nil || jwt.Validate(token) != nil {
		return fmt.Errorf("failed to validate token")
	}
	if jt.identifierField != "" {
		tMap, err := token.AsMap(context.Background())
		if err != nil {
			return fmt.Errorf("failed to parse token to map")
		}
		if tMap[jt.identifierField] == nil {
			return fmt.Errorf("failed to parse token map for field %s", jt.identifierField)
		}
	}
	return nil
}

// Authenticator is a default authentication middleware to enforce access from the
// Verifier middleware request context values. The Authenticator sends a 401 Unauthorized
// response for any unverified tokens and passes the good ones through. It's just fine
// until you decide to write something similar and customize your client response.
func (jt *JWT) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err := jt.CheckToken(token); err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

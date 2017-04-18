package auth

import (
	"errors"
	"fmt"
  "net/http"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io"
)

// accessToken
type Token string

// the key description as returned by the
// auth server
type keyDescription struct {
	Algorithm string `json:"algorithm" binding:"required"`
	Key       string `json:"key" binding:"required"`
}

// store public key in memory
var (
	algorithm string
	publicKey []byte
)

// fetch key from auth server
func fetchKey(server string) ([]byte, string, error) {
	// check if publicKey is set
	if publicKey == nil {
		uri := fmt.Sprintf("%s/key", server)

		resp, err := http.Get(uri)
		if err != nil {
			return nil, "", err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, "", errors.New(resp.Status)
		}

		// io.Copy(os.Stdout, resp.Body)

		decoder := json.NewDecoder(resp.Body)
		var k keyDescription
		if err := decoder.Decode(&k); err != nil {
			println(err)
			return nil, "", err
		}

		publicKey = []byte(k.Key)
		algorithm = k.Algorithm
	}

	return publicKey, algorithm, nil
}

// fetch the key from server for a specific token, making sure the algorithm is
// correct
func fetchKeyForToken(server string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		if key, algo, err := fetchKey(server); err != nil {
			return nil, err
		} else if algo != token.Header["alg"] {
			return nil, errors.New("invalid token algorithm")
		} else {
			return key, nil
		}
	}
}

// validate an access token
func ValidateToken (server, accessToken string) (map[string]interface{}, error) {

	token, err := jwt.Parse(accessToken, fetchKeyForToken(server))

	switch {
		case err != nil:
			return nil, err

		case !token.Valid:
			return nil, errors.New("token is invalid")

		default:
			exp := token.Claims["exp"]
			println(exp)
			return token.Claims, nil
	}
}


// middleware for routes that require authorization
func RequireAuth(server string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := jwt.ParseFromRequest(c.Request, fetchKeyForToken(server))

		switch {
			case err != nil:
				c.AbortWithError(401, err)
        //
			case !token.Valid:
				c.AbortWithError(401, errors.New("token is invalid"))

			default:
				// save token
				c.Set("token",  Token(token.Raw))
				c.Set("claims", token.Claims)
		}
	}
}


func NewAuthRequest (accessToken Token, method string, uri string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", string(accessToken)))

	return req, nil
}

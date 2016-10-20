package security

import (
	"testing"
	"time"

	. "github.com/smartystreets/assertions"
)

var privKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEII97KXBANi9c3EjrYmjAGOqTG40yIBnIsGRHjHJpZp2ToAoGCCqGSM49
AwEHoUQDQgAEjpDmYI4+tNGyOncpxWKfPs8mirDYOft1TEC43DTCN5vCSfupyBS7
ZKgUUjg4E0Aq5SIJENqeRP3tTko8O3VZYQ==
-----END EC PRIVATE KEY-----`

var pubKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEjpDmYI4+tNGyOncpxWKfPs8mirDY
Oft1TEC43DTCN5vCSfupyBS7ZKgUUjg4E0Aq5SIJENqeRP3tTko8O3VZYQ==
-----END PUBLIC KEY-----`

func TestJWT(t *testing.T) {
	a := New(t)

	jwt, err := BuildJWT("the-subject", time.Second, []byte(privKey))
	a.So(err, ShouldBeNil)

	claims, err := ValidateJWT(jwt, []byte(pubKey))
	a.So(err, ShouldBeNil)

	a.So(claims.Subject, ShouldEqual, "the-subject")
	a.So(claims.Issuer, ShouldEqual, "the-subject")

	// Wrong private key
	_, err = ValidateJWT(jwt, []byte("this is no key"))
	a.So(err, ShouldNotBeNil)

	// Wrong algorithm
	_, err = ValidateJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.Rq8IxqeX7eA6GgYxlcHdPFVRNFFZc5rEI3MQTZZbK3I", []byte(pubKey))
	a.So(err, ShouldNotBeNil)

	// Wrong signature
	_, err = ValidateJWT("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NzY4MDE0MDUsImlhdCI6MTQ3NjgwMTMyNSwiaXNzIjoidGhlLXN1YmplY3QiLCJuYmYiOjE0NzY4MDEzMjUsInN1YiI6InRoZS1zdWJqZWN0In0.4YudFUVQL4ODy8MMHWZrdB3CCgZedCD7FMUu2iPF4O1WIvptKaUyp9lBu-Eo2SfuNXcTIa1CiOiye36aeelCEw", []byte(pubKey))
	a.So(err, ShouldNotBeNil)

	// Expired
	_, err = ValidateJWT("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NzY4MDE1NzEsImlhdCI6MTQ3NjgwMTU1MCwiaXNzIjoidGhlLXN1YmplY3QiLCJuYmYiOjE0NzY4MDE1NTAsInN1YiI6InRoZS1zdWJqZWN0In0.AsKUzs9kenfqmDEtLXvNk3Akf_dkfU-8Zy8brHRawsOr64LxA0Mfb2Ufxwzk0JQr5Rtigw2RAVGirFyZI5meBQ", []byte(pubKey))
	a.So(err, ShouldNotBeNil)
}

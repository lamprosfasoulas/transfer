package auth

import (
	"net/http"

	"github.com/lamprosfasoulas/transfer/pkg/storage"
)

// User is a struct used to communicate with the AuthProvider.
type User struct {
	UID 		string // User ID
	Username	string // Username
	Files 		*[]storage.FileInfo // Not Used
}

// AuthenticationResponse is used to relay back the results of
// the AuthProviders' methods.
type AuthenticationResponse struct {
	Success 	bool 
	User		*User
	Error 		error
	Message 	string
	JwtToken	string
}

// AuthProvider is the interface that allows users to login.
// AuthProvider also implements a JWTProvider that generates
// and validates the JWT tokens.
type AuthProvider interface {
	Authenticate(r *http.Request) 		AuthenticationResponse
	JWTProvider
}

// JWTProvider is used to generate and validate JWT tokens for each
// AuthProvider. 
type JWTProvider interface {
	//Token should contain username
	GenerateToken(r AuthenticationResponse) AuthenticationResponse

	//Validate token should return AuthenticationResponse that 
	//contains username
	//This is used to get username for session during middleware
	//evaluation
	ValidateToken(token string)   			AuthenticationResponse
	//RefreshToken(refreshToken string) 	(string, error)
}

//func Authenticate(u, p string) (bool, error, string) {
//	var ok bool
//	var err error
//	var tokenString string
//
//	switch start.Cfg.AuthProvider {
//	case "LDAP":
//		ok, err = ldapAuth(u,p)
//		tokenString, err = token.GenerateJWT(u)
//	default:
//		return false, fmt.Errorf("No Provider specified"), ""
//	}
//	return ok, err, tokenString
//}

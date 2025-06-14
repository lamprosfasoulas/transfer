package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-ldap/ldap/v3"
	 "github.com/golang-jwt/jwt/v5"
)

// LdapProvider is the struct used 
// to communicate with the ldap service
type LdapProvider struct {
	LDAPURL			string
	LDAPBindDN		string
	LDAPBindPW		string
	LDAPBaseDN		string
	LDAPFilter		string
	*JWTLdap
}

// JWTLdap is used to store the information needed 
// for JWT token generation and validation.
// We make our own since ldap goes not provide JWT tokens.
type JWTLdap struct {
	Secret 			string
	Expiry 			time.Duration
}

// NewJWTLdap is the constructor for JWTLdap
func NewJWTLdap(secret string , expiry time.Duration) *JWTLdap{
	return &JWTLdap{
		Secret: secret,
		Expiry: expiry,
	}
}

// NewLdapProvider is the constructor for LdapProvider
func NewLdapProvider (url, bindDN, bindPW, baseDN, filter, jwtSec string, jwtExp time.Duration) *LdapProvider {
	return &LdapProvider{
		LDAPURL: 		url,
		LDAPBindDN:		bindDN,
		LDAPBindPW:		bindPW,
		LDAPBaseDN:		baseDN,
		LDAPFilter:		filter,
		JWTLdap: 		NewJWTLdap(
			jwtSec,	
			jwtExp,
			),
	}
}

// GenerateToken is used to generate new JWT tokens after Authentication.
func (J *JWTLdap) GenerateToken(r AuthenticationResponse) AuthenticationResponse {
	if r.Success {
		claims := jwt.MapClaims{
			"user": r.User.UID,
			"exp": time.Now().Add(J.Expiry).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		r.JwtToken, r.Error = token.SignedString([]byte(J.Secret))
		return r
	} else {
		return r
	}
}

// ValidateToken checks if the JWT is valid for used to have access.
func (J *JWTLdap) ValidateToken(t string) AuthenticationResponse {
	token, err := jwt.Parse(t, func(to *jwt.Token) (any, error) {
		if _, ok := to.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", to.Header["alg"])
		}
		return []byte(J.Secret), nil
	})
	if err != nil {
		return AuthenticationResponse{Success: false,Error: err}	
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if u, ok := claims["user"].(string); ok {
			return AuthenticationResponse{
				Success: true,
				User: &User{
					UID: u,
					Username: u,
				},
				JwtToken: t,
				Message: "JWT is OK",
			}
		}
		return AuthenticationResponse{
			Success: false,
			Message: "JWT does not contain user claim",
		}
	}

	return AuthenticationResponse{}
}

// Authenticate is the method used to check the user credentials against 
// the Provider. After successful check it responds with Success and the 
// JWTLdap generates the token the user need to get access to the pages.
func (L *LdapProvider) Authenticate (r *http.Request) AuthenticationResponse {
	if err := r.ParseForm(); err != nil {
		return AuthenticationResponse{
			Success: false,
			User: nil,
			Error: err,
			Message: "Failed to parse Form",
		}
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		return AuthenticationResponse{
			Success: false,
			User: nil,
			Error: fmt.Errorf("Username and Password are empty"),
			Message: "Username and Password are empty",
		}
	}
	user := &User{
		UID: username,
		Username: username,
	}

	ok, err := ldapAuth(username, password, L)
	if err != nil {
		return AuthenticationResponse{
			Success: ok,
			User: user,
			Error: err,
			Message: "Ldap auth error",
			//JWTToken: "",
		}
	}
	//token, err := generateJWT(username)
	//if err != nil {
	//	return AuthenticationResponse{
	//		Success: ok,
	//		User: user,
	//		Error: err,
	//		Message: "Ldap auth error",
	//		JWTToken: "",
	//	}
	//}
	return AuthenticationResponse{
		Success: ok,
		User: user,
		Error: err,
		Message: "Ldap auth complete",
		//JWTToken: token,
	}
}

// ldapAuth is an internal method that connect with the ldap server to check
// if the user exists and if his credentials are right.
func ldapAuth(username, password string, L *LdapProvider) (bool, error)  {
	l, err := ldap.DialURL(L.LDAPURL)
	if err != nil {
		return false, fmt.Errorf("Failed to connect to LDAP server: %v\n",err)
	}
	defer l.Close()

	err = l.Bind(L.LDAPBindDN, L.LDAPBindPW)
	if err != nil {
		return false, fmt.Errorf("Failed to bind with user: %v\n",err)
	}

	searchRequest := ldap.NewSearchRequest(
		L.LDAPBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(L.LDAPFilter, username),
		[]string{"dn","uid","cn","mail"},
		nil,
		)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return false, fmt.Errorf("Failed to search LDAP: %v\n",err)
	}
	if len (sr.Entries) != 1 {
		return false, fmt.Errorf("User not found or too many entries\n")
	}
	userDN := sr.Entries[0].DN
	err = l.Bind(userDN, password)
	if err != nil {
		return false, nil
	}
	return true, nil
}

//func (J *JWTLdap) GenerateJWT(username string) (string, error) {
//	claims := jwt.MapClaims{
//		"user": username,
//		"exp": time.Now().Add(J.Expiry).Unix(),
//		"iat": time.Now().Unix(),
//	}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	return token.SignedString([]byte(J.Secret))
//}

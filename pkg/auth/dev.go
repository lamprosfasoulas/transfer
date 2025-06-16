package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Develop struct {
	Username string
	Password string
	*JWTDev
}

type JWTDev struct {
	Secret string
	Expiry time.Duration
}

func NewJWTDev(s string, e time.Duration) *JWTDev {
	return &JWTDev{
		Secret: s,
		Expiry: e,
	}
}

func NewDevProvider(u, p, jwtSec string, jwtExp time.Duration) *Develop {
	return &Develop{
		Username: u,
		Password: p,
		JWTDev: NewJWTDev(
			jwtSec,
			jwtExp,
			),
	}
}

func (d *Develop) GenerateToken(r AuthenticationResponse) AuthenticationResponse {
	if r.Success {
		claims := jwt.MapClaims{
			"user": r.User.UID,
			"exp": time.Now().Add(d.Expiry).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		r.JwtToken, r.Error = token.SignedString([]byte(d.Secret))
		return r
	} else {
		return r
	}
}

// ValidateToken checks if the JWT is valid for used to have access.
func (d *Develop) ValidateToken(t string) AuthenticationResponse {
	token, err := jwt.Parse(t, func(to *jwt.Token) (any, error) {
		if _, ok := to.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", to.Header["alg"])
		}
		return []byte(d.Secret), nil
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


func (d *Develop) Authenticate(r *http.Request) AuthenticationResponse{
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
	
	if username == d.Username && password == d.Password {
		return AuthenticationResponse{
			Success: true,
			User: user,
			Error: nil,
			Message: "Dev login successful",
			//JWTToken: token,
		}
	}
	return AuthenticationResponse{
		Success: false,
		User: user,
		Error: nil,
		Message: "Dev login failed",
		//JWTToken: token,
	}
}

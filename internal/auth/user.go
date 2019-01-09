package auth

import (
	"errors"
	"fmt"

	"github.com/glynternet/mon/internal/jwt"
	"github.com/glynternet/mon/internal/model"
)

// UserClaimsAuthoriser authorises the claims of a JWT against the given user.
func UserClaimsAuthoriser(u model.User) jwt.ClaimsAuthoriser {
	return userClaimsAuthoriser{authorisedUser: u}
}

type userClaimsAuthoriser struct {
	authorisedUser model.User
}

// NewClaims provides an empty &model.User{} that the claims should me
// unmarshalled into.
func (userClaimsAuthoriser) NewClaims() interface{} {
	return &model.User{}
}

// Authorise returns an error if the given claims cannot be authorised against
// the Authoriser's authrosied user.
func (a userClaimsAuthoriser) Authorise(claims interface{}) error {
	u, ok := claims.(*model.User)
	if !ok {
		return fmt.Errorf("expected claims to be of type '*model.User' but got %T", claims)
	}
	if u == nil {
		return errors.New("nil user")
	}
	if a.authorisedUser.Email != u.Email {
		return errors.New("invalid email address")
	}
	if a.authorisedUser.EmailVerified && !u.EmailVerified {
		return errors.New("authorisation requires validated email address, received unverified email")
	}
	return nil
}
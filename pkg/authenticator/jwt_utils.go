package authenticator

import (
	"ichi-go/pkg/requestctx"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func ExtractToken(requestCtx requestctx.RequestContext) (string, error) {
	parts := strings.Split(requestCtx.Authorization, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", jwt.ErrTokenMalformed
	}
	return parts[1], nil
}

func ParseToken(tokenString string, jwtConfig JWTConfig) (*jwt.Token, error) {
	var options []jwt.ParserOption

	if jwtConfig.LeewayDuration > 0 {
		options = append(options, jwt.WithLeeway(jwtConfig.LeewayDuration))
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if ok != true {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtConfig.PublicKey, nil
	}, options...)

	if token == nil {
		return nil, jwt.ErrTokenMalformed
	}
	if err != nil {
		return nil, jwt.ErrInvalidKey
	}
	return token, nil
}

func ValidateToken(token jwt.Token) (jwt.Claims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func GetUserId(claims jwt.Claims) (UserSubject, error) {
	subj, er := claims.GetSubject()
	us := UserSubject{ID: 0}
	if er != nil {
		return us, jwt.ErrTokenInvalidSubject
	}
	userId, uError := strconv.ParseUint(subj, 10, 64)
	if uError != nil {
		return us, jwt.ErrTokenInvalidSubject
	}
	us.ID = userId
	return us, nil
}

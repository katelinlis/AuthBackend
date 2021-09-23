package model

import (
	"testing"
)

//TestJWT ...
func TestJWT(t *testing.T) *JWT {

	return &JWT{
		TokenString: "dfdfsfdfds",
		UserID:      1,
	}
}

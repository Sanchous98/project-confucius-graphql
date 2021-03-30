package graphql

import (
	"fmt"
	"github.com/graphql-go/graphql"
)

func IsGranted(field *graphql.Field, args map[string]interface{}) {
	field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		if !grantAccess() {
			return nil, fmt.Errorf("I have no access to this resource")
		}

		result, err := field.Resolve(p)

		if err != nil {
			return result, err
		}

		return make(map[string]interface{}), nil
	}
}

func grantAccess() bool {
	// TODO: Check rights before resolving
	return true
}

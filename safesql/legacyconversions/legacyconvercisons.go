package legacyconversions

import (
	"github.com/empijei/def-prog-exercises/safesql"
	"github.com/empijei/def-prog-exercises/safesql/internal/raw"
)

var trustedSQLCtor = raw.TrustedSQLCtor.(func(string) safesql.TrustedSQL)

/*** RiskilyAssumeTrustedSQL
* This function is used to convert a string to a TrustedSQL object.
* This is a function used as a middlestep to convert a string to a TrustedSQL object.
* When you got time, you should replace the usage of this function with the New function from safesql package.
***/
func RiskilyAssumeTrustedSQL(trusted string) safesql.TrustedSQL {
	return trustedSQLCtor(trusted)
}

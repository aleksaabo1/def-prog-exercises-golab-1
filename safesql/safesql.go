package safesql

import "github.com/empijei/def-prog-exercises/safesql/internal/raw"

type CompileTimeConstant string

type TrustedSQL struct {
	s string
}

func New(text CompileTimeConstant) TrustedSQL {
	return TrustedSQL{string(text)}
}

func init() {
	raw.TrustedSQLCtor = func(unsafe string) TrustedSQL {
		return TrustedSQL{unsafe}
	}
}

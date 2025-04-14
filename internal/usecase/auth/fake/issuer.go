package fake

type Issuer struct {
	Token string
	Err   error
}

func (fi *Issuer) Issue(_ string) (string, error) {
	return fi.Token, fi.Err
}

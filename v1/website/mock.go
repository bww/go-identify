package website

import (
	"context"
)

// mockResolver always produces the info it's provided
type mockResolver struct {
	Info Info
	Err  error
}

// NewMock produces a new resolver that always returns the provided info and error
func NewMock(info Info, err error) Resolver {
	return &mockResolver{
		Info: info,
		Err:  err,
	}
}

func (r *mockResolver) IdentifyDomain(cxt context.Context, domain string) (Info, error) {
	return r.Info, r.Err
}

func (r *mockResolver) IdentifyWebsite(cxt context.Context, link string) (Info, error) {
	return r.Info, r.Err
}

package website

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMock(t *testing.T) {
	einfo := Info{Owner: "Test Enterprises, Inc"}
	eerr := errors.New("Test error")
	r := NewMock(einfo, eerr)

	rinfo, rerr := r.IdentifyDomain(nil, "google.com")
	assert.Equal(t, einfo, rinfo)
	assert.Equal(t, eerr, rerr)

	rinfo, rerr = r.IdentifyWebsite(nil, "https://google.com")
	assert.Equal(t, einfo, rinfo)
	assert.Equal(t, eerr, rerr)
}

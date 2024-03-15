package aliaser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestPackagesErrors(t *testing.T) {
	err1 := packages.Error{Msg: "error 1"}
	err2 := packages.Error{Msg: "error 2"}
	errs := PackagesErrors{err1, err2}
	assert.Equal(t, err1.Error()+"; "+err2.Error(), errs.Error())
}

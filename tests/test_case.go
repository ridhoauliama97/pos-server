package tests

import (
	"github.com/goravel/framework/testing"

	"github.com/ridhoauliama97/pos-server/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}

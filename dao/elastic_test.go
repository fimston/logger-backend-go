package dao

import (
	"testing"
)

func TestAliasCreate(test *testing.T) {
	el := NewElastic("http://localhost:9200")
	err := el.CreateAccount(NewAccountInfo(1))
	if err != nil {
		test.Fatal(err)
	}
}

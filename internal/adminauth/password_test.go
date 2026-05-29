package adminauth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("12345678")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "12345678") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected password mismatch")
	}
}

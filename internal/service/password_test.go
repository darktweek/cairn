package service

import "testing"

func TestPasswordRoundtrip(t *testing.T) {
	hash, err := hashPassword("corr3ct-horse")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !verifyPassword("corr3ct-horse", hash) {
		t.Fatal("correct password must verify")
	}
	if verifyPassword("wrong-password", hash) {
		t.Fatal("wrong password must not verify")
	}
}

func TestVerifyMalformedHash(t *testing.T) {
	for _, h := range []string{
		"",
		"plaintext",
		"$argon2id$v=19$m=65536,t=1,p=4$only-five-parts",
		"$argon2id$v=19$m=65536,t=1,p=4$!!!notbase64!!!$AAAA",
	} {
		if verifyPassword("anything", h) {
			t.Fatalf("malformed hash %q must not verify", h)
		}
	}
}

func TestHashesAreSalted(t *testing.T) {
	a, _ := hashPassword("same")
	b, _ := hashPassword("same")
	if a == b {
		t.Fatal("two hashes of the same password must differ (random salt)")
	}
}

// internal/types/cardhash_test.go
package types

import (
	"testing"
)

// TestHashBytesKnownValue matches Rust's test_display:
// blake3("test") == "4878ca0425c739fa427f7eda20fe845f6b2e46ba5fe2a14df5b1e32f50603215"
func TestHashBytesKnownValue(t *testing.T) {
	hash := HashBytes([]byte("test"))
	want := "4878ca0425c739fa427f7eda20fe845f6b2e46ba5fe2a14df5b1e32f50603215"
	if hash.String() != want {
		t.Errorf("HashBytes(\"test\") = %q, want %q", hash.String(), want)
	}
}

// TestCardHashOrdering matches Rust's test_ordering.
func TestCardHashOrdering(t *testing.T) {
	a, err := ParseCardHash("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ParseCardHash("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseCardHash("0000000000000000000000000000000000000000000000000000000000000002")
	if err != nil {
		t.Fatal(err)
	}
	if !a.Less(b) {
		t.Error("a < b failed")
	}
	if !b.Less(c) {
		t.Error("b < c failed")
	}
}

// TestParseCardHashRoundtrip verifies Hex → ParseCardHash roundtrip.
func TestParseCardHashRoundtrip(t *testing.T) {
	original := HashBytes([]byte("hello world"))
	hex := original.Hex()
	parsed, err := ParseCardHash(hex)
	if err != nil {
		t.Fatalf("ParseCardHash(%q): %v", hex, err)
	}
	if !original.Equal(parsed) {
		t.Errorf("roundtrip failed: %q != %q", original.Hex(), parsed.Hex())
	}
}

// TestParseCardHashInvalid verifies that invalid hex strings return an error.
func TestParseCardHashInvalid(t *testing.T) {
	invalids := []string{
		"",
		"notahex",
		"4878ca0425c739fa427f7eda20fe845f6b2e46ba5fe2a14df5b1e32f5060321",   // 63 chars (too short)
		"4878ca0425c739fa427f7eda20fe845f6b2e46ba5fe2a14df5b1e32f506032150", // 65 chars (too long)
	}
	for _, s := range invalids {
		if _, err := ParseCardHash(s); err == nil {
			t.Errorf("ParseCardHash(%q): expected error, got nil", s)
		}
	}
}

// TestBasicCardHash matches Rust's test_basic_card_hash.
func TestBasicCardHash(t *testing.T) {
	c1 := NewBasicContent("What is 2+2?", "4")
	c2 := NewBasicContent("What is 2+2?", "4")
	c3 := NewBasicContent("What is 3+3?", "6")
	if !c1.Hash().Equal(c2.Hash()) {
		t.Error("same content should produce equal hashes")
	}
	if c1.Hash().Equal(c3.Hash()) {
		t.Error("different content should produce different hashes")
	}
}

// TestClozeCardHash matches Rust's test_cloze_card_hash.
func TestClozeCardHash(t *testing.T) {
	// Two cloze cards from the same text share the same family hash but
	// have different card hashes.
	a := NewClozeContent("The capital of France is Paris", 0, 1)
	b := NewClozeContent("The capital of France is Paris", 0, 2)
	if a.Hash().Equal(b.Hash()) {
		t.Error("cloze cards with different end positions should have different hashes")
	}
	if !(*a.FamilyHash()).Equal(*b.FamilyHash()) {
		t.Error("cloze cards from the same text should share a family hash")
	}
}

// TestFamilyHashNilForBasic verifies that basic cards return nil family hash.
func TestFamilyHashNilForBasic(t *testing.T) {
	c := NewBasicContent("Q", "A")
	if c.FamilyHash() != nil {
		t.Error("basic card should return nil family hash")
	}
}

// TestHasherIncremental verifies that the Hasher accumulates data correctly
// and produces the same result as a single HashBytes call for the same content.
func TestHasherIncremental(t *testing.T) {
	// Basic card hash is computed as: blake3("Basic" + question + answer)
	h := NewHasher()
	h.Update([]byte("Basic"))
	h.Update([]byte("Hello"))
	h.Update([]byte("World"))
	want := h.Finalize()

	c := NewBasicContent("Hello", "World")
	got := c.Hash()
	if !got.Equal(want) {
		t.Errorf("incremental hash mismatch: got %s, want %s", got.Hex(), want.Hex())
	}
}

package rng

import (
	"testing"
)

// TestNextU32Deterministic verifies that the same seed always produces the
// same sequence, matching the requirement in the Rust implementation.
func TestNextU32Deterministic(t *testing.T) {
	r1 := FromSeed(42)
	r2 := FromSeed(42)

	for i := 0; i < 100; i++ {
		v1 := r1.NextU32()
		v2 := r2.NextU32()
		if v1 != v2 {
			t.Errorf("step %d: NextU32() = %d vs %d (not deterministic)", i, v1, v2)
		}
	}
}

// TestDifferentSeedsProduceDifferentOutput verifies that distinct seeds diverge.
func TestDifferentSeedsProduceDifferentOutput(t *testing.T) {
	r1 := FromSeed(1)
	r2 := FromSeed(2)

	allEqual := true
	for i := 0; i < 10; i++ {
		if r1.NextU32() != r2.NextU32() {
			allEqual = false
			break
		}
	}
	if allEqual {
		t.Error("different seeds should not produce identical sequences")
	}
}

// TestGenerateRange verifies Generate(max) always returns a value in [0, max).
func TestGenerateRange(t *testing.T) {
	r := FromSeed(12345)
	for i := 0; i < 1000; i++ {
		for _, max := range []uint32{1, 2, 10, 100, 1000} {
			got := r.Generate(max)
			if got >= max {
				t.Errorf("Generate(%d) = %d, want < %d", max, got, max)
			}
		}
	}
}

// TestShuffleDeterministic verifies that Shuffle with the same seed produces
// the same permutation every time.
func TestShuffleDeterministic(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	r1 := FromSeed(999)
	r2 := FromSeed(999)
	out1 := Shuffle(input, r1)
	out2 := Shuffle(input, r2)

	for i := range out1 {
		if out1[i] != out2[i] {
			t.Errorf("position %d: Shuffle produced %d vs %d", i, out1[i], out2[i])
		}
	}
}

// TestShufflePreservesElements verifies that Shuffle does not lose or
// duplicate any element.
func TestShufflePreservesElements(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	r := FromSeed(777)
	output := Shuffle(input, r)

	if len(output) != len(input) {
		t.Fatalf("len(output) = %d, want %d", len(output), len(input))
	}
	counts := make(map[int]int)
	for _, v := range input {
		counts[v]++
	}
	for _, v := range output {
		counts[v]--
	}
	for v, c := range counts {
		if c != 0 {
			t.Errorf("element %d count diff = %d (input and output differ)", v, c)
		}
	}
}

// TestShuffleDoesNotMutateInput verifies that the original slice is not modified.
func TestShuffleDoesNotMutateInput(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	original := make([]int, len(input))
	copy(original, input)

	r := FromSeed(42)
	Shuffle(input, r)

	for i := range input {
		if input[i] != original[i] {
			t.Errorf("input[%d] was modified: got %d, want %d", i, input[i], original[i])
		}
	}
}

// TestLCGConstants verifies the LCG constants match the Rust implementation
// exactly, ensuring cross-language RNG compatibility.
func TestLCGConstants(t *testing.T) {
	const wantA uint64 = 6364136223846793005
	const wantC uint64 = 1442695040888963407
	if lcgA != wantA {
		t.Errorf("lcgA = %d, want %d", lcgA, wantA)
	}
	if lcgC != wantC {
		t.Errorf("lcgC = %d, want %d", lcgC, wantC)
	}
}

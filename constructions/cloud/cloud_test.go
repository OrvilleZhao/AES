package cloud

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"

	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/saes"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var (
	key   = []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}
	seed  = []byte{38, 41, 142, 156, 29, 181, 23, 194, 21, 250, 223, 183, 210, 168, 214, 145}
	input = []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
)

func TestSubBytes(t *testing.T) {
	baseConstr := saes.Construction{make([]byte, 16)}

	// Generate a random test vector.
	vect := make([]byte, 16)
	rand.Read(vect)

	// Calculate the correct result with StandardAES.
	real := make([]byte, 16)
	copy(real, vect)
	baseConstr.SubBytes(real)

	// Calculate our result with manual inversions / matrix math.
	cand := make([]byte, 16)
	copy(cand, vect)

	for pos := 0; pos < 16; pos++ {
		cand[pos] = Invert{}.Get(cand[pos])
	}

	copy(cand, SubBytes.Mul(matrix.Row(cand)))

	for pos := 0; pos < 16; pos++ {
		cand[pos] ^= SubBytesConst[pos]
	}

	// Check that the two are equal.
	if bytes.Compare(real, cand) != 0 {
		t.Fatalf("SubBytes matrix was wrong for input: %x\nReal: %x\nCand: %x", vect, real, cand)
	}
}

func TestMixColumns(t *testing.T) {
	baseConstr := saes.Construction{make([]byte, 16)}

	vect := make([]byte, 16)
	rand.Read(vect)

	real := make([]byte, 16)
	copy(real, vect)
	baseConstr.MixColumns(real)

	cand := make([]byte, 16)
	copy(cand, MixColumns.Mul(matrix.Row(vect)))

	if bytes.Compare(real, cand) != 0 {
		t.Fatalf("MixColumns matrix was wrong for input: %x\nReal: %x\nCand: %x", vect, real, cand)
	}
}

func TestShiftRows(t *testing.T) {
	baseConstr := saes.Construction{make([]byte, 16)}

	vect := make([]byte, 16)
	rand.Read(vect)

	real := make([]byte, 16)
	copy(real, vect)
	baseConstr.ShiftRows(real)

	cand := make([]byte, 16)
	copy(cand, ShiftRows.Mul(matrix.Row(vect)))

	if bytes.Compare(real, cand) != 0 {
		t.Fatalf("ShiftRows matrix was wrong for input: %x\nReal: %x\nCand: %x", vect, real, cand)
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors {
		constr, inputMask, outputMask := GenerateEncryptionKeys(
			vec.Key, vec.Key, common.IndependentMasks{common.RandomMask, common.RandomMask},
		)

		inputInv, _ := inputMask.Invert()
		outputInv, _ := outputMask.Invert()

		in, out := make([]byte, 16), make([]byte, 16)

		copy(in, inputInv.Mul(matrix.Row(vec.In))) // Apply input encoding.

		constr.Encrypt(out, in)

		copy(out, outputInv.Mul(matrix.Row(out))) // Remove output encoding.

		if !bytes.Equal(vec.Out, out) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, out)
		}
	}
}

func TestPersistence(t *testing.T) {
	constr1, _, _ := GenerateEncryptionKeys(key, seed, common.IndependentMasks{common.RandomMask, common.RandomMask})

	serialized := constr1.Serialize()
	constr2, err := Parse(serialized)

	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	cand1, cand2 := make([]byte, 16), make([]byte, 16)

	constr1.Encrypt(cand1, input)
	constr2.Encrypt(cand2, input)

	if !bytes.Equal(cand1, cand2) {
		t.Fatalf("Real disagrees with parsed! %x != %x", cand1, cand2)
	}
}
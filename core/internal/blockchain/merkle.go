package blockchain

import "crypto/sha256"

// MerkleTree holds the leaves and provides root/proof operations.
type MerkleTree struct {
	Leaves [][]byte
}

// hashPair returns SHA256(left || right).
func hashPair(a, b []byte) []byte {
	h := sha256.New()
	h.Write(a)
	h.Write(b)
	return h.Sum(nil)
}

// Root computes the Merkle root from the leaves.
// If there is only one leaf, its hash is the root.
// Odd layers are padded by duplicating the last element.
func (t *MerkleTree) Root() []byte {
	if len(t.Leaves) == 0 {
		return make([]byte, 32) // zero hash
	}

	level := make([][]byte, len(t.Leaves))
	copy(level, t.Leaves)

	for len(level) > 1 {
		if len(level)%2 != 0 {
			level = append(level, level[len(level)-1]) // duplicate last
		}
		next := make([][]byte, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			next[i/2] = hashPair(level[i], level[i+1])
		}
		level = next
	}
	return level[0]
}

// Proof generates a Merkle proof for the leaf at the given index.
// Returns the sibling hashes and their positions (false=left, true=right).
type ProofStep struct {
	Hash    []byte `json:"hash"`
	IsRight bool   `json:"isRight"` // true if this sibling is on the right
}

func (t *MerkleTree) Proof(index int) []ProofStep {
	if index < 0 || index >= len(t.Leaves) {
		return nil
	}

	level := make([][]byte, len(t.Leaves))
	copy(level, t.Leaves)

	var proof []ProofStep
	idx := index

	for len(level) > 1 {
		if len(level)%2 != 0 {
			level = append(level, level[len(level)-1])
		}

		// Sibling index
		var sibIdx int
		if idx%2 == 0 {
			sibIdx = idx + 1
			proof = append(proof, ProofStep{Hash: level[sibIdx], IsRight: true})
		} else {
			sibIdx = idx - 1
			proof = append(proof, ProofStep{Hash: level[sibIdx], IsRight: false})
		}
		_ = sibIdx

		// Move to parent level
		next := make([][]byte, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			next[i/2] = hashPair(level[i], level[i+1])
		}
		level = next
		idx /= 2
	}
	return proof
}

// VerifyProof checks that a leaf + proof reconstructs the expected root.
func VerifyProof(leaf, root []byte, proof []ProofStep) bool {
	current := leaf
	for _, step := range proof {
		if step.IsRight {
			current = hashPair(current, step.Hash)
		} else {
			current = hashPair(step.Hash, current)
		}
	}
	if len(current) != len(root) {
		return false
	}
	for i := range current {
		if current[i] != root[i] {
			return false
		}
	}
	return true
}

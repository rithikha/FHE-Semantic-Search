package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

var (
	params    ckks.Parameters
	sk        *rlwe.SecretKey
	pk        *rlwe.PublicKey
	encoder   ckks.Encoder
	evaluator ckks.Evaluator
)

func initLattigo() {
    var err error
    params, err = ckks.NewParametersFromLiteral(ckks.PN13QP218)
    if err != nil {
        panic(err)
    }

    keygen := ckks.NewKeyGenerator(params)
    sk, pk = keygen.GenKeyPair()

    encoder = ckks.NewEncoder(params)

    // Generate required rotations for summation
    rotations := []int{}
    slots := params.Slots()
    for i := 1; i < int(slots); i <<= 1 {
        rotations = append(rotations, i)
    }

    rotKeys := keygen.GenRotationKeysForRotations(rotations, false, sk)
    rlk := keygen.GenRelinearizationKey(sk, 1)

    // Initialize evaluator with the generated keys
    evaluator = ckks.NewEvaluator(params, rlwe.EvaluationKey{Rlk: rlk, Rtks: rotKeys})
}


// Fetch ciphertext embedding from Iroh by CID
func fetchEmbeddingFromIroh(cid string) ([]byte, error) {
	tmpFile := "retrieved_embedding.bin"
	cmd := exec.Command("iroh", "blobs", "get", cid, "-o", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, err
	}
	os.Remove(tmpFile)
	return data, nil
}

// Deserialize ciphertext from bytes
func ciphertextFromBytes(data []byte) (*rlwe.Ciphertext, error) {
	ct := new(rlwe.Ciphertext)
	err := ct.UnmarshalBinary(data)
	return ct, err
}

// Perform homomorphic dot product (similarity)
func homomorphicDotProduct(queryCT, entryCT *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	// Element-wise multiply ciphertexts
	product := evaluator.MulNew(queryCT, entryCT)
	evaluator.Relinearize(product, product)

	// Sum across ciphertext slots homomorphically
	slots := params.Slots()
	for offset := 1; offset < int(slots); offset *= 2 {
		rotated := evaluator.RotateNew(product, offset)
		product = evaluator.AddNew(product, rotated)
	}
	return product, nil
}

func main() {
	initLattigo()

	// Replace explicitly with your actual CID from frontend logs
	entryCID := "blobacw4v3kesexwxjdokonxgx7zvm7g6qgo2s3bch5pv33nbqvsvjoniajdnb2hi4dthixs6ylqomys2mjoojswyylzfzuxe33ifzxgk5dxn5zgwlrpaeaau3wix7cfoadf54vq5hcomafg2bl3a5t5mbasjxihfx4vjo2ut7ioqo4x4gpaf4"

	entryCipherBytes, err := fetchEmbeddingFromIroh(entryCID)
	if err != nil {
		panic(fmt.Sprintf("Iroh fetch error: %v", err))
	}

	entryCiphertext, err := ciphertextFromBytes(entryCipherBytes)
	if err != nil {
		panic(fmt.Sprintf("Ciphertext decode error: %v", err))
	}

	queryCipherBytes, err := os.ReadFile("encrypted_query.bin")
	if err != nil {
		panic(fmt.Sprintf("Error reading encrypted_query.bin: %v", err))
	}

	queryCiphertext, err := ciphertextFromBytes(queryCipherBytes)
	if err != nil {
		panic(fmt.Sprintf("Query ciphertext decode error: %v", err))
	}

	// Perform homomorphic similarity computation
	resultCiphertext, err := homomorphicDotProduct(queryCiphertext, entryCiphertext)
	if err != nil {
		panic(fmt.Sprintf("Homomorphic computation error: %v", err))
	}

	resultBytes, err := resultCiphertext.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("Ciphertext serialize error: %v", err))
	}

	resultBase64 := base64.StdEncoding.EncodeToString(resultBytes)
	fmt.Println("Encrypted similarity score (base64):", resultBase64)

	err = os.WriteFile("encrypted_similarity_result.bin", resultBytes, 0644)
	if err != nil {
		panic(fmt.Sprintf("File write error: %v", err))
	}

	fmt.Println("Encrypted similarity computation completed successfully. Result saved to 'encrypted_similarity_result.bin'.")
}

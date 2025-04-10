package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tuneinsight/lattigo/v5/core/rlwe"
	"github.com/tuneinsight/lattigo/v5/schemes/ckks"
)

var (
	params    ckks.Parameters
	encoder   *ckks.Encoder
	decryptor *rlwe.Decryptor
	evaluator *ckks.Evaluator
	sk        *rlwe.SecretKey
	pk        *rlwe.PublicKey
)

type EntryEmbedding struct {
	Name string `json:"name"`
	CID  string `json:"cid"`
}

var entryEmbeddings map[string]EntryEmbedding

func initLattigo() {
	var err error
	params, err = ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN: 13,
		LogQ: []int{60, 40, 40, 60},
		LogP: []int{60},
	})
	if err != nil {
		panic("Failed to initialize CKKS params: " + err.Error())
	}

	keygen := ckks.NewKeyGenerator(params)
	sk, pk = keygen.GenKeyPairNew()

	encoder = ckks.NewEncoder(params)
	decryptor = rlwe.NewDecryptor(params, sk)

	var rotations []uint64
	slots := params.MaxSlots()
	for i := uint64(1); i < slots; i <<= 1 {
		rotations = append(rotations, i)
	}

	rotKeys := keygen.GenGaloisKeysNew(rotations, sk)
	rlk := keygen.GenRelinearizationKeyNew(sk)

	evaluator = ckks.NewEvaluator(params, rlwe.NewMemEvaluationKeySet(rlk, rotKeys...))

	fmt.Println("CKKS setup complete with slots:", slots)
}

func fetchEmbeddingFromIroh(ticket string) ([]byte, error) {
	tmpFile := fmt.Sprintf("tmp_embedding_%s.bin", ticket)
	cmd := exec.Command("iroh", "blobs", "get", "--ticket", ticket, "-o", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(tmpFile)
	os.Remove(tmpFile)
	return data, err
}

func ciphertextFromBytes(data []byte) (*rlwe.Ciphertext, error) {
	ct := new(rlwe.Ciphertext)
	err := ct.UnmarshalBinary(data)
	return ct, err
}

func homomorphicDotProduct(queryCT, entryCT *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	fmt.Println("Computing encrypted dot product...")

	product, err := evaluator.MulNew(queryCT, entryCT)
	if err != nil {
		return nil, fmt.Errorf("Mul error: %v", err)
	}

	if err := evaluator.Relinearize(product, product); err != nil {
		return nil, fmt.Errorf("Relinearize error: %v", err)
	}

	if err := evaluator.Rescale(product, product); err != nil {
		return nil, fmt.Errorf("Rescale error: %v", err)
	}

	// Rotate + Add to sum all slots
	slots := params.MaxSlots()
	for offset := 1; offset < int(slots); offset <<= 1 {
		rotated, err := evaluator.RotateNew(product, offset)
		if err != nil {
			return nil, fmt.Errorf("Rotate error: %v", err)
		}
		evaluator.Add(product, rotated, product)
	}

	return product, nil
}

func getEncryptedSimilarity(c *gin.Context) {
	queryCID := c.Query("cid")
	if queryCID == "" {
		c.JSON(400, gin.H{"error": "Missing query CID parameter"})
		return
	}

	queryCipherBytes, err := fetchEmbeddingFromIroh(queryCID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Query fetch error: " + err.Error()})
		return
	}

	queryCiphertext, err := ciphertextFromBytes(queryCipherBytes)
	if err != nil {
		c.JSON(500, gin.H{"error": "Ciphertext decode error: " + err.Error()})
		return
	}

	var encryptedResults []map[string]interface{}
	for entryID, entry := range entryEmbeddings {
		fmt.Println("Matching against:", entry.Name)

		entryCipherBytes, err := fetchEmbeddingFromIroh(entry.CID)
		if err != nil {
			log.Printf("Iroh fetch error for %s: %v", entry.Name, err)
			continue
		}
		entryCiphertext, err := ciphertextFromBytes(entryCipherBytes)
		if err != nil {
			log.Printf("Cipher decode error for %s: %v", entry.Name, err)
			continue
		}

		resultCiphertext, err := homomorphicDotProduct(queryCiphertext, entryCiphertext)
		if err != nil {
			log.Printf("Homomorphic error for %s: %v", entry.Name, err)
			continue
		}

		cipherBytes, err := resultCiphertext.MarshalBinary()
		if err != nil {
			log.Printf("Marshal error for %s: %v", entry.Name, err)
			continue
		}

		encryptedResults = append(encryptedResults, map[string]interface{}{
			"entry_id":     entryID,
			"name":         entry.Name,
			"score_cipher": base64.StdEncoding.EncodeToString(cipherBytes),
		})
	}

	c.JSON(200, encryptedResults)
}

func getPublicKey(c *gin.Context) {
	pkBytes, _ := pk.MarshalBinary()
	c.JSON(200, gin.H{"public_key": base64.StdEncoding.EncodeToString(pkBytes)})
}

func main() {
	initLattigo()

	// Load manifest
	manifestData, err := os.ReadFile("embeddings_manifest.json")
	if err != nil {
		panic("Failed to read manifest: " + err.Error())
	}
	var manifest struct {
		V3 struct {
			EntryEmbeddings map[string]EntryEmbedding `json:"entry_embeddings"`
		} `json:"v3"`
	}
	json.Unmarshal(manifestData, &manifest)
	entryEmbeddings = manifest.V3.EntryEmbeddings

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/get_encrypted_similarity", getEncryptedSimilarity)
	router.GET("/public_key", getPublicKey)

	fmt.Println("FHE server live at http://localhost:5000")
	router.Run(":5000")
}

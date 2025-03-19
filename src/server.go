package main

import (
	"fmt"
	"os"
	"os/exec"
)

func fetchEmbeddingFromIroh(cid string) ([]byte, error) {
	tmpFile := "retrieved_embedding.bin"

	cmd := exec.Command("iroh", "blobs", "get", cid, "-o", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Iroh fetch error: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, err
	}

	os.Remove(tmpFile)
	return data, nil
}

func main() {
	cid := "blobacw4v3kesexwxjdokonxgx7zvm7g6qgo2s3bch5pv33nbqvsvjoniajdnb2hi4dthixs6ylqomys2mjoojswyylzfzuxe33ifzxgk5dxn5zgwlrpaeaau3wix7cfoadf54vq5hcomafg2bl3a5t5mbasjxihfx4vjo2ut7ioqo4x4gpaf4"
	data, err := fetchEmbeddingFromIroh(cid)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully retrieved ciphertext embedding (%d bytes)\n", len(data))
}

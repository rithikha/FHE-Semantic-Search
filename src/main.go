package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

var (
	params    ckks.Parameters
	pk        *rlwe.PublicKey
	encryptor rlwe.Encryptor
	encoder   ckks.Encoder
)

// Initialization of CKKS parameters, keys, encryptor, and encoder.
func init() {
	var err error
	params, err = ckks.NewParametersFromLiteral(ckks.PN13QP218)
	if err != nil {
		panic(err)
	}

	keygen := ckks.NewKeyGenerator(params)
	_, pk = keygen.GenKeyPair()

	encryptor = ckks.NewEncryptor(params, pk)
	encoder = ckks.NewEncoder(params)
}

// encryptVector encrypts a JSON-formatted vector received from JavaScript.
func encryptVector(this js.Value, args []js.Value) interface{} {
	vectorJSON := args[0].String()

	var vector []float64
	if err := json.Unmarshal([]byte(vectorJSON), &vector); err != nil {
		return js.ValueOf(err.Error())
	}

	complexVector := make([]complex128, len(vector))
	for i, val := range vector {
		complexVector[i] = complex(val, 0)
	}

	plaintext := encoder.EncodeNew(complexVector, params.MaxLevel(), params.DefaultScale(), params.LogSlots())
	ciphertext := encryptor.EncryptNew(plaintext)

	cipherData, err := ciphertext.MarshalBinary()
	if err != nil {
		return js.ValueOf(err.Error())
	}

	arrayConstructor := js.Global().Get("Uint8Array")
	jsCipherArray := arrayConstructor.New(len(cipherData))
	js.CopyBytesToJS(jsCipherArray, cipherData)

	return jsCipherArray
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("encryptVector", js.FuncOf(encryptVector))
	println("Lattigo WASM module loaded successfully.")
	<-c
}

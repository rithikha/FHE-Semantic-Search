# Privacy-Preserving Semantic Search [In-Progress]

FSS is a prototype demonstrating a privacy-preserving semantic search system using Fully Homomorphic Encryption (FHE), specifically the CKKS scheme via Lattigo, combined with decentralized storage via Iroh. This project aims to match user queries with directory entries based on semantic similarity without revealing either the queries or the directory content.


## How It Works

1. **Directory Preparation:**
   - Each directory entry (e.g., user profiles, interests) is transformed into embeddings using Sentence-BERT embeddings.
   - Embeddings are normalized and encrypted using CKKS via Lattigo.
   - Encrypted embeddings are uploaded and stored securely on Iroh, a decentralized storage solution.

2. **Query Processing:**
   - User inputs a plaintext query in the browser.
   - Query is transformed into embeddings using Xenova's transformer model directly in-browser.
   - Embeddings are encrypted client-side via WebAssembly-compiled Go code.
   - Encrypted query embeddings are uploaded to Iroh.

3. **Encrypted Semantic Matching:**
   - The Go backend retrieves the encrypted query embedding from Iroh.
   - It performs homomorphic dot-product computations between the query and each directory embedding, producing encrypted similarity scores without decrypting data at any point.

4. **Results:**
   - The encrypted similarity scores are returned to the client.
   - Optionally, the client decrypts and ranks these scores locally.


# Running the FSS Project

### 1. Install Dependencies

Make sure you have **Go 1.21+**, **Iroh CLI** installed, and your browser supports **WebAssembly**.

Run this command to install Go dependencies:

```bash
go mod tidy
```

---

### 2. Prepare Directory Embeddings

- Populate `directory_entries.json` with your directory entries.
- Generate normalized embeddings for each entry and save them in the `embeddings/` folder.

Then, run:

```bash
go run prep_directory.go
```

This generates an `embeddings_manifest.json` after encrypting the embeddings and uploading them to Iroh.

---

### 3. Start the Backend Server

Run this command to start the backend:

```bash
go run server.go
```

Your server will start running at:

```
http://localhost:5000
```

---

### 4. Frontend Querying

- Open `index.html` in your web browser.
- Enter plaintext queries that will be automatically transformed, encrypted, and securely matched with directory entries.

---

## 🛠 Technologies Used

- **Frontend:** HTML + WebAssembly
- **Embeddings:** Xenova transformers (browser-side Sentence-BERT)
- **Encryption:** CKKS via Lattigo
- **Storage:** Iroh decentralized storage
- **Backend:** Go, Gin web framework

---

## TODO / Future Improvements 

- Multi-party integration 
- Matching only if both parties consent
- Integration of Zero-Knowledge proofs (e.g., Noir) for additional verification of skillset, credentials, etc

---

## Thank You's

Many people to thank in the making of this project: Micah, Ayush, Vivek, Anka, Emma and the rest of the CPR team and friends


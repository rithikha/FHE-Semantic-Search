# generate the embeddings from directory download/import
from sentence_transformers import SentenceTransformer
import json
import numpy as np
import os

# Load the pretrained Sentence-BERT model
model = SentenceTransformer('all-MiniLM-L6-v2')

# Load directory entries
with open('directory_entries.json', 'r') as file:
    entries = json.load(file)

# Extract textual interests from each entry
texts = [entry['interests'] for entry in entries]

# Generate embeddings - vector representing interests
embeddings = model.encode(texts, batch_size=32, show_progress_bar=True)
# Normalize embeddings to unit length for future cosine similarity computation
embeddings_norm = embeddings / np.linalg.norm(embeddings, axis=1, keepdims=True)


# increment version after each update
version = 1

# check that embeddings directory exists
os.makedirs("embeddings", exist_ok=True)

# Save embeddings
np.save(f"embeddings/embeddings_v{version}.npy", embeddings_norm)

# Save IDs alongside embeddings for later retrieval
ids = [entry["id"] for entry in entries]
with open(f"embeddings/indices_v{version}.json", "w") as f:
    json.dump({"ids": ids}, f)

print(f"Embeddings v{version} generated and saved successfully.")
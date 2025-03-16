import numpy as np
import json
from sentence_transformers import util

# Load embeddings
embeddings = np.load('src/embeddings/embeddings_v2.npy')

# Check shape (number of entries, vector dimensions)
print("Embeddings shape:", embeddings.shape)

# Inspect first embedding vector
print("First embedding vector:", embeddings[0])

# Inspect first few embedding vectors
print("First 5 embeddings:")
print(embeddings[:5])

# Semantic similarity inspection

# Load the original textual entries for context
with open('src/directory_entries.json', 'r') as file:
    entries = json.load(file)

texts = [entry['interests'] for entry in entries]

print("\nOriginal texts corresponding to embeddings:")
for idx, text in enumerate(texts):
    print(f"{idx}: {text}")

# Semantic similarity (cosine similarity) clearly printed
print("\nSemantic similarity to first entry:")
similarities = util.cos_sim(embeddings[0], embeddings)[0].numpy()

for idx, score in enumerate(similarities):
    print(f"Similarity between entry 0 and entry {idx} ('{texts[idx]}'): {score:.4f}")

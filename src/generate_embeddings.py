import os
import json
import numpy as np

from sentence_transformers import SentenceTransformer
from datetime import datetime

# Config
entries_file = 'directory_entries.json'  
embeddings_dir = 'embeddings'            
manifest_file = 'embeddings_manifest.json'

# Load Sentence-BERT model
model = SentenceTransformer('all-MiniLM-L6-v2')

# Load existing entries
with open(entries_file, 'r') as file:
    entries = json.load(file)

# Extract textual interests for embeddings 
texts = [entry['interests'] for entry in entries]

# Generate embeddings 
embeddings = model.encode(texts, show_progress_bar=True)

# Normalize embeddings
embeddings_norm = embeddings / np.linalg.norm(embeddings, axis=1, keepdims=True)

# Ensure embeddings directory exists
os.makedirs(embeddings_dir, exist_ok=True)

# Automate versioning
if os.path.exists(manifest_file):
    with open(manifest_file, 'r') as f:
        manifest = json.load(f)
        latest_version = manifest.get('latest_version', 0)
else:
    manifest = {}
    latest_version = 0

version = latest_version + 1  # automatically increment version number 

# Save normalized embeddings file (exactly as in tutorial, but automated version number)
embeddings_file = os.path.join(embeddings_dir, f'embeddings_v{version}.npy')
np.save(embeddings_file, embeddings_norm)

# Save indices mapping file 
indices_file = os.path.join(embeddings_dir, f'indices_v{version}.json')
ids = [entry['id'] for entry in entries]
with open(indices_file, 'w') as f:
    json.dump({"ids": ids}, f, indent=2)

# Update manifest file 
# Load existing manifest/start fresh
manifest['latest_version'] = version
manifest[f'v{version}'] = {
    'embeddings_file': embeddings_file,
    'indices_file': indices_file,
    'timestamp': datetime.now().isoformat()
}

# Save updated manifest 
with open(manifest_file, 'w') as f:
    json.dump(manifest, f, indent=2)

# print confirmation of embeddings update
print(f"Embeddings version {version} generated, normalized, and saved successfully.")
print(f"Manifest updated: embeddings_manifest.json")

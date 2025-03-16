import os
import json
import numpy as np
import subprocess

from sentence_transformers import SentenceTransformer
from datetime import datetime

# Config
entries_file = 'directory_entries.json'  
embeddings_dir = 'src/embeddings'
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

# Load existing manifest or start fresh
if os.path.exists(manifest_file):
    with open(manifest_file, 'r') as f:
        manifest = json.load(f)
        latest_version = manifest.get('latest_version', 0)
else:
    manifest = {}
    latest_version = 0

version = latest_version + 1  # automatically increment version number 

# Save normalized embeddings file
embeddings_file = os.path.join(embeddings_dir, f'embeddings_v{version}.npy')
np.save(embeddings_file, embeddings / np.linalg.norm(embeddings, axis=1, keepdims=True))

# Save indices file
indices_file = os.path.join(embeddings_dir, f'indices_v{version}.json')
ids = [entry['id'] for entry in entries]
with open(indices_file, 'w') as f:
    json.dump({"ids": ids}, f, indent=2)

# Add embeddings file to Iroh (automatically)
print(f"Adding embeddings file to Iroh: {embeddings_file}")
result = subprocess.run(
    ['iroh', 'blobs', 'add', embeddings_file],
    stdout=subprocess.PIPE,
    text=True
)

# Parse CID and Ticket from Iroh output
iroh_output = result.stdout.strip().splitlines()
cid = None
ticket = None
for line in iroh_output:
    if line.startswith('Blob:'):
        cid = line.split('Blob: ')[1].strip()
    elif line.startswith('All-in-one ticket'):
        ticket = line.split(': ', 1)[1].strip()

if not cid or not ticket:
    raise ValueError("Failed to parse CID or Ticket from Iroh output.")

# Update manifest file 
manifest['latest_version'] = version
manifest[f'v{version}'] = {
    'embeddings_file': embeddings_file,
    'indices_file': indices_file,
    'iroh_blob_cid': cid,
    'iroh_ticket': ticket,
    'timestamp': datetime.now().isoformat()
}

# Save updated manifest file
with open(manifest_file, 'w') as f:
    json.dump(manifest, f, indent=2)

# Print confirmation of embeddings update
print(f"Embeddings version {version} generated, normalized, saved locally, and added to Iroh successfully.")
print(f"Manifest updated: {manifest_file}")
print(f"Iroh CID: {cid}")
print(f"Iroh Ticket: {ticket}")

# Automatically run the cleanup script at the very end of your generate_embeddings.py
print("🔄 Running cleanup_embeddings.py automatically...")
subprocess.run(["python", "cleanup_embeddings.py"])
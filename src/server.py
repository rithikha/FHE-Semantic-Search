from flask import Flask, request, jsonify, send_from_directory
from flask_cors import CORS
from sentence_transformers import SentenceTransformer
import numpy as np
import json

# Initialize Flask app
app = Flask(__name__)
CORS(app)

# Load embeddings manifest once at startup
with open('embeddings/embeddings_manifest.json', 'r') as f:
    manifest = json.load(f)

latest_version = manifest["latest_version"]
embeddings_file = manifest[f"v{latest_version}"]["embeddings_file"]
indices_file = manifest[f"v{latest_version}"]["indices_file"]

# Load embeddings and entry IDs
embeddings = np.load(embeddings_file)
with open(indices_file, 'r') as f:
    entry_ids = json.load(f)["ids"]

# Normalize embeddings for efficient similarity search
embeddings = embeddings / np.linalg.norm(embeddings, axis=1, keepdims=True)

# Load directory entries once at startup
with open('directory_entries.json', 'r') as f:
    entries = json.load(f)

# Initialize SentenceTransformer model once
model = SentenceTransformer('all-MiniLM-L6-v2')

@app.route('/search', methods=['GET'])
def search():
    user_query = request.args.get('q', default="", type=str)
    if not user_query:
        return jsonify({"error": "No query provided"}), 400

    # Encode and normalize the user's query
    query_vec = model.encode(user_query)
    query_vec /= np.linalg.norm(query_vec)

    # Compute cosine similarities with stored embeddings
    similarities = np.dot(embeddings, query_vec)

    # Get top 3 closest matches
    top_k = 3
    top_indices = similarities.argsort()[::-1][:top_k]

    results = []
    for idx in top_indices:
        entry = entries[idx]
        results.append({
            "id": entry_ids[idx],
            "name": entry["name"],
            "interests": entry["interests"],
            "score": round(float(similarities[idx]), 4)
        })

    return jsonify({"query": user_query, "results": results})

@app.route('/')
def serve_frontend():
    return send_from_directory('.', 'index.html')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
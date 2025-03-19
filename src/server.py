from flask import Flask, request, jsonify
from flask_cors import CORS
import subprocess
import tempfile
import os

app = Flask(__name__)
CORS(app)

@app.route('/store_embedding', methods=['POST'])
def store_embedding():
    encrypted_embedding = request.data  # Get raw binary data from request

    if not encrypted_embedding:
        return jsonify({"error": "No data received"}), 400

    try:
        with tempfile.NamedTemporaryFile(delete=False) as tmp_file:
            tmp_file.write(encrypted_embedding)
            tmp_file_path = tmp_file.name

        # Store embedding in Iroh using CLI
        result = subprocess.run(
            ["iroh", "blobs", "add", tmp_file_path],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            return jsonify({"error": "Iroh storage failed", "details": result.stderr}), 500

        cid = result.stdout.strip()  # Iroh returns CID here

        # Clean up temporary file
        os.remove(tmp_file_path)

        # Return CID to frontend
        return jsonify({"cid": cid})

    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(port=5000, debug=True)

import os
import json
import subprocess

# Configuration
EMBEDDINGS_MANIFEST = 'embeddings_manifest.json'
MAX_VERSIONS_TO_KEEP = 2

# Load manifest
if not os.path.exists(EMBEDDINGS_MANIFEST):
    print("Manifest not found. Nothing to cleanup.")
    exit(1)

with open(EMBEDDINGS_MANIFEST, 'r') as f:
    manifest = json.load(f)

# Identify all versions
versions = sorted(
    [int(key[1:]) for key in manifest if key.startswith('v')],
    reverse=True
)

# Only keep the newest N versions
versions_to_delete = versions[MAX_VERSIONS_TO_KEEP:]

# Cleanup older versions
for version in versions_to_delete:
    key = f'v{version}'
    embeddings_file = manifest[key]['embeddings_file']
    indices_file = manifest[key]['indices_file']
    cid = manifest[key].get('iroh_blob_cid')

    # Delete local embedding file
    if os.path.exists(embeddings_file):
        os.remove(embeddings_file)
        print(f"Removed local file: {embeddings_file}")

    # Delete indices file
    if os.path.exists(indices_file):
        os.remove(indices_file)
        print(f"Removed indices file: {indices_file}")

    # Remove blob from Iroh (if CID is stored)
    if cid := manifest[key].get('iroh_blob_cid'):
        subprocess.run(["iroh", "blobs", "delete", cid])
        print(f"Removed Iroh blob: {cid}")

    # Remove version from manifest
    del manifest[key]

# Save updated manifest
manifest['latest_version'] = max(versions[:MAX_VERSIONS_TO_KEEP])
with open(EMBEDDINGS_MANIFEST, 'w') as f:
    json.dump(manifest, f, indent=2)

print(f"Cleaned up. Kept latest {MAX_VERSIONS_TO_KEEP} versions.")

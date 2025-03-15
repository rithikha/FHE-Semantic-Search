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

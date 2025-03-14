// Import the keccak256 hash function from Noir's standard library
use std::hash::keccak256;

fn main(interest_hash: pub [u8; 32], secret_interest: [u8; 8]) {
    // Compute the Keccak-256 hash of the secret interest (assuming 8-byte input)
    let computed_hash = keccak256(secret_interest, 8);
    // Assert that the computed hash matches the public hash provided
    assert(computed_hash == interest_hash);
}

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func main() {
	// Our AES-256 implementation requires a master key that is exactly 32 bytes long.
	// To make it printable and easy to copy-paste into the .env file, we generate
	// 24 secure random bytes and encode them using Base64.
	// 24 bytes * (4/3) = exactly 32 Base64 characters.
	
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Fatalf("Failed to generate secure random bytes: %v", err)
	}

	// Use URLEncoding to avoid special characters like '+' and '/' which can sometimes
	// cause formatting issues in shell environments.
	masterKey := base64.URLEncoding.EncodeToString(randomBytes)

	fmt.Println("=======================================================")
	fmt.Println("            Dboke Master Key Generator                 ")
	fmt.Println("=======================================================")
	fmt.Println("\nGenerated 32-character AES-256 Master Key:")
	fmt.Printf("\n    %s\n\n", masterKey)
	fmt.Println("Instructions:")
	fmt.Println("1. Copy the key above.")
	fmt.Println("2. Open your '.env' file in the root directory.")
	fmt.Println("3. Replace DBOKE_MASTER_KEY with this new value.")
	fmt.Println("4. DO NOT commit this key to version control.")
	fmt.Println("5. KEEP IT SECRET! If lost, saved DB passwords cannot be recovered.")
	fmt.Println("\n=======================================================")
}

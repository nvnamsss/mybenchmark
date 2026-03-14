package mybenchmark

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Helper function to generate HMAC secret key
func generateHMACSecret() []byte {
	secret := make([]byte, 64)
	rand.Read(secret)
	return secret
}

// Helper function to generate RSA key pair
func generateRSAKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// Helper function to generate ECDSA key pair
func generateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// Helper function to generate Ed25519 key pair
func generateEd25519KeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	return priv, pub, err
}

// Helper function to create a standard token with minimal claims
func createStandardClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"sub": "1234567890",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
}

// ============================================================================
// Signing (Token Generation) Benchmarks
// ============================================================================

func BenchmarkJWTSign_HS256(b *testing.B) {
	secret := generateHMACSecret()
	claims := createStandardClaims()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		_, err := token.SignedString(secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTSign_HS512(b *testing.B) {
	secret := generateHMACSecret()
	claims := createStandardClaims()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		_, err := token.SignedString(secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTSign_RS256(b *testing.B) {
	privateKey, err := generateRSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	claims := createStandardClaims()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		_, err := token.SignedString(privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTSign_PS256(b *testing.B) {
	privateKey, err := generateRSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	claims := createStandardClaims()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
		_, err := token.SignedString(privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTSign_ES256(b *testing.B) {
	privateKey, err := generateECDSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	claims := createStandardClaims()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		_, err := token.SignedString(privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTSign_EdDSA(b *testing.B) {
	privateKey, _, err := generateEd25519KeyPair()
	if err != nil {
		b.Fatal(err)
	}
	claims := createStandardClaims()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		_, err := token.SignedString(privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// Verification Benchmarks (Serial)
// ============================================================================

func BenchmarkJWTVerify_HS256(b *testing.B) {
	secret := generateHMACSecret()
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTVerify_HS512(b *testing.B) {
	secret := generateHMACSecret()
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTVerify_RS256(b *testing.B) {
	privateKey, err := generateRSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTVerify_PS256(b *testing.B) {
	privateKey, err := generateRSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTVerify_ES256(b *testing.B) {
	privateKey, err := generateECDSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTVerify_EdDSA(b *testing.B) {
	privateKey, publicKey, err := generateEd25519KeyPair()
	if err != nil {
		b.Fatal(err)
	}
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// Verification Benchmarks (Parallel)
// ============================================================================

func BenchmarkJWTVerify_HS256_Parallel(b *testing.B) {
	secret := generateHMACSecret()
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkJWTVerify_HS512_Parallel(b *testing.B) {
	secret := generateHMACSecret()
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkJWTVerify_RS256_Parallel(b *testing.B) {
	privateKey, err := generateRSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return publicKey, nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkJWTVerify_PS256_Parallel(b *testing.B) {
	privateKey, err := generateRSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return publicKey, nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkJWTVerify_ES256_Parallel(b *testing.B) {
	privateKey, err := generateECDSAKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return publicKey, nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkJWTVerify_EdDSA_Parallel(b *testing.B) {
	privateKey, publicKey, err := generateEd25519KeyPair()
	if err != nil {
		b.Fatal(err)
	}
	claims := createStandardClaims()

	// Pre-generate token
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return publicKey, nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

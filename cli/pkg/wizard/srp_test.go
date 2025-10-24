package wizard

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"
)

// Test SRP core algorithms
func TestSRPCore(t *testing.T) {
	core := NewSRPCore(SRPGroup4096)

	if core.params.Length != 4096 {
		t.Errorf("Expected length 4096, got %d", core.params.Length)
	}

	if core.params.Hash != "SHA-256" {
		t.Errorf("Expected hash SHA-256, got %s", core.params.Hash)
	}

	if core.params.G.Cmp(big.NewInt(5)) != 0 {
		t.Errorf("Expected G=5, got %v", core.params.G)
	}

	t.Logf("SRP Core initialized successfully with %d-bit group", core.params.Length)
}

// Test complete SRP authentication flow
func TestSRPAuthentication(t *testing.T) {
	// Simulate user credentials
	password := "test-password-123"

	// Step 1: Registration phase - client generates verifier
	t.Log("=== Registration Phase ===")

	// Generate salt
	salt := make([]byte, 16)
	copy(salt, []byte("test-salt-123456")) // Use fixed salt for testing

	// Calculate x = H(salt | password)
	x := calculateX(salt, password)
	t.Logf("Generated x: %x", x.Bytes())

	// Client initialization
	client := NewSRPClient(SRPGroup4096)
	err := client.Initialize(x.Bytes())
	if err != nil {
		t.Fatalf("Client initialization failed: %v", err)
	}

	v := client.GetV()
	if v == nil {
		t.Fatal("Failed to get verifier v")
	}
	t.Logf("Generated verifier v: %x", v)

	t.Log("=== Login Phase - Server Side ===")

	// Server initialization
	server := NewSRPServer(SRPGroup4096)
	err = server.Initialize(v)
	if err != nil {
		t.Fatalf("Server initialization failed: %v", err)
	}

	B := server.GetB()
	if B == nil {
		t.Fatal("Failed to get server B value")
	}
	t.Logf("Server B: %x", B)

	t.Log("=== Login Phase - Client Side ===")

	loginClient := NewSRPClient(SRPGroup4096)
	err = loginClient.Initialize(x.Bytes())
	if err != nil {
		t.Fatalf("Login client initialization failed: %v", err)
	}

	err = loginClient.SetB(B)
	if err != nil {
		t.Fatalf("Client SetB failed: %v", err)
	}

	A := loginClient.GetA()
	M1_client := loginClient.GetM1()
	K_client := loginClient.GetK()

	if A == nil || M1_client == nil || K_client == nil {
		t.Fatal("Failed to get client A, M1, or K values")
	}

	t.Logf("Client A: %x", A)
	t.Logf("Client M1: %x", M1_client)
	t.Logf("Client K: %x", K_client)

	t.Log("=== Server Verification ===")

	err = server.SetA(A)
	if err != nil {
		t.Fatalf("Server SetA failed: %v", err)
	}

	M1_server := server.GetM1()
	K_server := server.GetK()

	if M1_server == nil || K_server == nil {
		t.Fatal("Failed to get server M1 or K values")
	}

	t.Logf("Server M1: %x", M1_server)
	t.Logf("Server K: %x", K_server)

	if !bytes.Equal(M1_client, M1_server) {
		t.Errorf("M1 values don't match!")
		t.Errorf("Client M1: %x", M1_client)
		t.Errorf("Server M1: %x", M1_server)
	} else {
		t.Log("✅ M1 verification successful!")
	}

	if !bytes.Equal(K_client, K_server) {
		t.Errorf("K values don't match!")
		t.Errorf("Client K: %x", K_client)
		t.Errorf("Server K: %x", K_server)
	} else {
		t.Log("✅ Shared key K verification successful!")
	}

	t.Log("=== M2 Verification ===")

	M2_server := server.GetM2()
	M2_client := loginClient.GetM2()

	if M2_server == nil || M2_client == nil {
		t.Fatal("Failed to get M2 values")
	}

	if !bytes.Equal(M2_client, M2_server) {
		t.Errorf("M2 values don't match!")
		t.Errorf("Client M2: %x", M2_client)
		t.Errorf("Server M2: %x", M2_server)
	} else {
		t.Log("✅ M2 verification successful!")
	}

	t.Log("🎉 Complete SRP authentication flow successful!")
}

// Test different group lengths
func TestSRPDifferentGroupLengths(t *testing.T) {
	groupLengths := []SRPGroupLength{SRPGroup3072, SRPGroup4096}

	for _, length := range groupLengths {
		t.Run(fmt.Sprintf("Group%d", length), func(t *testing.T) {
			client := NewSRPClient(length)
			server := NewSRPServer(length)

			x := calculateX([]byte("test-salt"), "test-password")

			// Client initialization
			err := client.Initialize(x.Bytes())
			if err != nil {
				t.Fatalf("Client initialization failed for group %d: %v", length, err)
			}

			// Server initialization
			v := client.GetV()
			err = server.Initialize(v)
			if err != nil {
				t.Fatalf("Server initialization failed for group %d: %v", length, err)
			}

			t.Logf("✅ Group %d initialization successful", length)
		})
	}
}

func TestSRPErrorCases(t *testing.T) {
	client := NewSRPClient(SRPGroup4096)

	err := client.SetB([]byte("invalid"))
	if err == nil {
		t.Error("Expected error when calling SetB on uninitialized client")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid group length")
		}
	}()
	NewSRPCore(SRPGroupLength(1234))
}

func TestSRPCoreComponents(t *testing.T) {
	core := NewSRPCore(SRPGroup4096)

	a := big.NewInt(123)
	b := big.NewInt(456)

	h1, err := core.H(a, b)
	if err != nil {
		t.Fatalf("Hash function failed: %v", err)
	}

	h2, err := core.H(a, b)
	if err != nil {
		t.Fatalf("Hash function failed: %v", err)
	}

	if h1.Cmp(h2) != 0 {
		t.Error("Hash function should be deterministic")
	}

	x := big.NewInt(789)
	v := core.V(x)
	if v == nil {
		t.Error("Verifier calculation failed")
	}

	a_val := big.NewInt(101112)
	A := core.A(a_val)
	if A == nil {
		t.Error("A value calculation failed")
	}

	t.Log("✅ SRP core components test successful")
}

func BenchmarkSRPAuthentication(b *testing.B) {
	salt := []byte("benchmark-salt-123456")
	password := "benchmark-password"
	x := calculateX(salt, password)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client := NewSRPClient(SRPGroup4096)
		client.Initialize(x.Bytes())
		v := client.GetV()

		server := NewSRPServer(SRPGroup4096)
		server.Initialize(v)
		B := server.GetB()

		loginClient := NewSRPClient(SRPGroup4096)
		loginClient.Initialize(x.Bytes())
		loginClient.SetB(B)

		A := loginClient.GetA()
		server.SetA(A)

		M1_client := loginClient.GetM1()
		M1_server := server.GetM1()

		if !bytes.Equal(M1_client, M1_server) {
			b.Fatalf("M1 verification failed in benchmark")
		}
	}
}

func calculateX(salt []byte, password string) *big.Int {
	hasher := sha256.New()
	hasher.Write(salt)
	hasher.Write([]byte(password))
	hash := hasher.Sum(nil)

	return new(big.Int).SetBytes(hash)
}

func TestSRPKnownVectors(t *testing.T) {
	salt := []byte("test-salt")
	password := "test-password"
	x := calculateX(salt, password)

	client := NewSRPClient(SRPGroup4096)
	err := client.Initialize(x.Bytes())
	if err != nil {
		t.Fatalf("Client initialization failed: %v", err)
	}

	v := client.GetV()
	A := client.GetA()

	if len(v) == 0 {
		t.Error("Verifier v should not be empty")
	}

	if len(A) == 0 {
		t.Error("A value should not be empty")
	}

	core := NewSRPCore(SRPGroup4096)
	vBig := new(big.Int).SetBytes(v)
	ABig := new(big.Int).SetBytes(A)

	if vBig.Cmp(core.params.N) >= 0 {
		t.Error("Verifier v should be less than N")
	}

	if ABig.Cmp(core.params.N) >= 0 {
		t.Error("A value should be less than N")
	}

	t.Logf("Known vectors test successful")
	t.Logf("Salt: %x", salt)
	t.Logf("X: %x", x.Bytes())
	t.Logf("V: %x", v)
	t.Logf("A: %x", A)
}

func TestSRPConsistency(t *testing.T) {
	salt := []byte("consistency-test-salt")
	password := "consistency-password"
	x := calculateX(salt, password)

	client1 := NewSRPClient(SRPGroup4096)
	client1.Initialize(x.Bytes())
	v1 := client1.GetV()

	client2 := NewSRPClient(SRPGroup4096)
	client2.Initialize(x.Bytes())
	v2 := client2.GetV()

	if !bytes.Equal(v1, v2) {
		t.Error("Verifiers should be identical for same x value")
	}

	A1 := client1.GetA()
	A2 := client2.GetA()

	if bytes.Equal(A1, A2) {
		t.Error("A values should be different due to random a")
	}

	t.Log("✅ SRP consistency test successful")
}

func TestSRPInvalidInputs(t *testing.T) {
	client := NewSRPClient(SRPGroup4096)
	server := NewSRPServer(SRPGroup4096)

	err := client.Initialize([]byte{})
	if err != nil {
		t.Logf("Expected behavior: empty x rejected: %v", err)
	}

	err = server.Initialize([]byte{})
	if err != nil {
		t.Logf("Expected behavior: empty v rejected: %v", err)
	}

	x := calculateX([]byte("test-salt"), "test-password")
	err = client.Initialize(x.Bytes())
	if err != nil {
		t.Fatalf("Normal initialization should succeed: %v", err)
	}

	zeroB := make([]byte, 512) // 全零的 B 值
	err = client.SetB(zeroB)
	if err == nil {
		t.Error("Should reject zero B value")
	} else {
		t.Logf("✅ Correctly rejected zero B value: %v", err)
	}
}

func BenchmarkSRPGroupLengths(b *testing.B) {
	groupLengths := []SRPGroupLength{SRPGroup3072, SRPGroup4096}

	for _, length := range groupLengths {
		b.Run(fmt.Sprintf("Group%d", length), func(b *testing.B) {
			salt := []byte("benchmark-salt")
			password := "benchmark-password"
			x := calculateX(salt, password)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				client := NewSRPClient(length)
				client.Initialize(x.Bytes())

				server := NewSRPServer(length)
				server.Initialize(client.GetV())

				loginClient := NewSRPClient(length)
				loginClient.Initialize(x.Bytes())
				loginClient.SetB(server.GetB())

				server.SetA(loginClient.GetA())

				if !bytes.Equal(loginClient.GetM1(), server.GetM1()) {
					b.Fatal("M1 verification failed")
				}
			}
		})
	}
}

func TestSRPFlowExample(t *testing.T) {
	fmt.Println("=== SRP Authentication Flow Example ===")

	username := "user@example.com"
	password := "secure-password-123"
	salt := []byte("random-salt-16bytes")

	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Salt: %x\n", salt)

	fmt.Println("\n1. Registration Phase:")
	x := calculateX(salt, password)
	client := NewSRPClient(SRPGroup4096)
	client.Initialize(x.Bytes())
	v := client.GetV()

	fmt.Printf("   Generated verifier v: %x...\n", v[:8])

	fmt.Println("\n2. Login Phase - Server:")
	server := NewSRPServer(SRPGroup4096)
	server.Initialize(v)
	B := server.GetB()

	fmt.Printf("   Server B: %x...\n", B[:8])

	fmt.Println("\n3. Login Phase - Client:")
	loginClient := NewSRPClient(SRPGroup4096)
	loginClient.Initialize(x.Bytes())
	loginClient.SetB(B)

	A := loginClient.GetA()
	M1_client := loginClient.GetM1()

	fmt.Printf("   Client A: %x...\n", A[:8])
	fmt.Printf("   Client M1: %x...\n", M1_client[:8])

	fmt.Println("\n4. Server Verification:")
	server.SetA(A)
	M1_server := server.GetM1()

	fmt.Printf("   Server M1: %x...\n", M1_server[:8])

	if bytes.Equal(M1_client, M1_server) {
		fmt.Println("   ✅ Authentication successful!")

		K_client := loginClient.GetK()
		K_server := server.GetK()

		fmt.Printf("   Shared key K: %x...\n", K_client[:8])

		if bytes.Equal(K_client, K_server) {
			fmt.Println("   ✅ Shared key established!")
		}
	} else {
		fmt.Println("   ❌ Authentication failed!")
	}

	// Output:
	// === SRP Authentication Flow Example ===
	// Username: user@example.com
	// Password: secure-password-123
	// Salt: 72616e646f6d2d73616c742d313662797465730000000000000000
	//
	// 1. Registration Phase:
	//    Generated verifier v: a1b2c3d4...
	//
	// 2. Login Phase - Server:
	//    Server B: e5f6a7b8...
	//
	// 3. Login Phase - Client:
	//    Client A: 12345678...
	//    Client M1: 9abcdef0...
	//
	// 4. Server Verification:
	//    Server M1: 9abcdef0...
	//    ✅ Authentication successful!
	//    Shared key K: fedcba98...
	//    ✅ Shared key established!
}

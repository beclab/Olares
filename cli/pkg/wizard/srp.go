package wizard

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"
)

// SRPGroupLength represents SRP group length types
type SRPGroupLength int

const (
	SRPGroup3072 SRPGroupLength = 3072
	SRPGroup4096 SRPGroupLength = 4096
	SRPGroup6144 SRPGroupLength = 6144
	SRPGroup8192 SRPGroupLength = 8192
)

// SRPParams represents SRP parameters
type SRPParams struct {
	Length SRPGroupLength
	Hash   string // "SHA-256"
	G      *big.Int
	N      *big.Int
}

// SRPSession represents SRP session state
type SRPSession struct {
	ID             string    `json:"id"`
	Created        time.Time `json:"created"`
	FailedAttempts int       `json:"failedAttempts"`
	AsAdmin        bool      `json:"asAdmin"`
	X              *big.Int  `json:"x,omitempty"`
	V              *big.Int  `json:"v,omitempty"`
	A              *big.Int  `json:"a,omitempty"`
	BigA           *big.Int  `json:"A,omitempty"`
	B              *big.Int  `json:"b,omitempty"`
	BigB           *big.Int  `json:"B,omitempty"`
	K              *big.Int  `json:"K,omitempty"`
	M1             *big.Int  `json:"M1,omitempty"`
	M2             *big.Int  `json:"M2,omitempty"`
}

func NewSRPSession() *SRPSession {
	return &SRPSession{
		ID:      generateUUID(),
		Created: time.Now(),
	}
}

// SRPClient represents SRP client
type SRPClient struct {
	session *SRPSession
	core    *SRPCore
}

func NewSRPClient(length SRPGroupLength) *SRPClient {
	return &SRPClient{
		session: NewSRPSession(),
		core:    NewSRPCore(length),
	}
}

// Getter methods
func (c *SRPClient) GetV() []byte {
	if c.session.V == nil {
		return nil
	}
	return c.session.V.Bytes()
}

func (c *SRPClient) GetA() []byte {
	if c.session.BigA == nil {
		return nil
	}
	return c.session.BigA.Bytes()
}

func (c *SRPClient) GetK() []byte {
	if c.session.K == nil {
		return nil
	}
	return c.session.K.Bytes()
}

func (c *SRPClient) GetM1() []byte {
	if c.session.M1 == nil {
		return nil
	}
	return c.session.M1.Bytes()
}

func (c *SRPClient) GetM2() []byte {
	if c.session.M2 == nil {
		return nil
	}
	return c.session.M2.Bytes()
}

// Initialize initializes SRP client
func (c *SRPClient) Initialize(x []byte) error {
	c.session.X = new(big.Int).SetBytes(x)
	c.session.V = c.core.V(c.session.X)
	
	// Generate random a
	a, err := generateRandomBigInt(32)
	if err != nil {
		return fmt.Errorf("failed to generate random a: %v", err)
	}
	c.session.A = a
	c.session.BigA = c.core.A(a)
	
	return nil
}

// SetB sets server's B value
func (c *SRPClient) SetB(B []byte) error {
	if c.session.X == nil || c.session.A == nil || c.session.BigA == nil {
		return fmt.Errorf("client not initialized")
	}
	
	c.session.BigB = new(big.Int).SetBytes(B)
	
	// Calculate K, M1, M2
	var err error
	c.session.K, err = c.getKey()
	if err != nil {
		return fmt.Errorf("failed to calculate key: %v", err)
	}
	
	c.session.M1, err = c.core.M1(c.session.BigA, c.session.BigB, c.session.K)
	if err != nil {
		return fmt.Errorf("failed to calculate M1: %v", err)
	}
	
	c.session.M2, err = c.core.M2(c.session.BigA, c.session.M1, c.session.K)
	if err != nil {
		return fmt.Errorf("failed to calculate M2: %v", err)
	}
	
	return nil
}

func (c *SRPClient) getKey() (*big.Int, error) {
	if c.session.X == nil || c.session.A == nil || c.session.BigA == nil || c.session.BigB == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	
	if c.core.IsZeroWhenModN(c.session.BigB) {
		return nil, fmt.Errorf("invalid B value")
	}
	
	u, err := c.core.U(c.session.BigA, c.session.BigB)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate u: %v", err)
	}
	
	S, err := c.core.ClientS(c.session.BigB, c.session.X, c.session.A, u)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate S: %v", err)
	}
	
	return c.core.K(S)
}

// SRPServer represents SRP server
type SRPServer struct {
	session *SRPSession
	core    *SRPCore
}

func NewSRPServer(length SRPGroupLength) *SRPServer {
	return &SRPServer{
		session: NewSRPSession(),
		core:    NewSRPCore(length),
	}
}

// Getter methods
func (s *SRPServer) GetB() []byte {
	if s.session.BigB == nil {
		return nil
	}
	return s.session.BigB.Bytes()
}

func (s *SRPServer) GetK() []byte {
	if s.session.K == nil {
		return nil
	}
	return s.session.K.Bytes()
}

func (s *SRPServer) GetM1() []byte {
	if s.session.M1 == nil {
		return nil
	}
	return s.session.M1.Bytes()
}

func (s *SRPServer) GetM2() []byte {
	if s.session.M2 == nil {
		return nil
	}
	return s.session.M2.Bytes()
}

// Initialize initializes SRP server
func (s *SRPServer) Initialize(v []byte) error {
	s.session.V = new(big.Int).SetBytes(v)
	
	// Generate random b
	b, err := generateRandomBigInt(32)
	if err != nil {
		return fmt.Errorf("failed to generate random b: %v", err)
	}
	s.session.B = b
	
	s.session.BigB, err = s.core.B(s.session.V, b)
	if err != nil {
		return fmt.Errorf("failed to calculate B: %v", err)
	}
	
	return nil
}

// SetA sets client's A value
func (s *SRPServer) SetA(A []byte) error {
	if s.session.V == nil || s.session.B == nil || s.session.BigB == nil {
		return fmt.Errorf("server not initialized")
	}
	
	s.session.BigA = new(big.Int).SetBytes(A)
	
	// Calculate K, M1, M2
	var err error
	s.session.K, err = s.getKey()
	if err != nil {
		return fmt.Errorf("failed to calculate key: %v", err)
	}
	
	s.session.M1, err = s.core.M1(s.session.BigA, s.session.BigB, s.session.K)
	if err != nil {
		return fmt.Errorf("failed to calculate M1: %v", err)
	}
	
	s.session.M2, err = s.core.M2(s.session.BigA, s.session.M1, s.session.K)
	if err != nil {
		return fmt.Errorf("failed to calculate M2: %v", err)
	}
	
	return nil
}

func (s *SRPServer) getKey() (*big.Int, error) {
	if s.session.V == nil || s.session.B == nil || s.session.BigB == nil || s.session.BigA == nil {
		return nil, fmt.Errorf("server not initialized")
	}
	
	if s.core.IsZeroWhenModN(s.session.BigA) {
		return nil, fmt.Errorf("invalid A value")
	}
	
	u, err := s.core.U(s.session.BigA, s.session.BigB)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate u: %v", err)
	}
	
	S := s.core.ServerS(s.session.BigA, s.session.V, u, s.session.B)
	return s.core.K(S)
}

// SRPCore implements core SRP algorithms
type SRPCore struct {
	params *SRPParams
}

func NewSRPCore(length SRPGroupLength) *SRPCore {
	params, err := getSRPParams(length)
	if err != nil {
		panic(fmt.Sprintf("Invalid group length: %v", err))
	}
	
	return &SRPCore{
		params: params,
	}
}

// H hash function (...inp) - ref: TypeScript srp.ts line 384-386
func (c *SRPCore) H(inputs ...*big.Int) (*big.Int, error) {
	hasher := sha256.New()
	
	for _, input := range inputs {
		// Ensure byte array length is correct (ref: TypeScript i2b function line 133-138)
		bytes := c.bigIntToBytes(input)
		hasher.Write(bytes)
	}
	
	hash := hasher.Sum(nil)
	return new(big.Int).SetBytes(hash), nil
}

// bigIntToBytes converts big integer to byte array, ensuring even length (ref: TypeScript i2b function)
func (c *SRPCore) bigIntToBytes(i *big.Int) []byte {
	hex := i.Text(16) // Convert to hexadecimal string
	
	// Ensure hexadecimal string has even length (ref: TypeScript srp.ts line 135-136)
	if len(hex)%2 == 1 {
		hex = "0" + hex
	}
	
	// Convert hexadecimal string to byte array
	bytes := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		var b byte
		fmt.Sscanf(hex[i:i+2], "%02x", &b)
		bytes[i/2] = b
	}
	
	return bytes
}

// V calculates verifier v = g^x % N
func (c *SRPCore) V(x *big.Int) *big.Int {
	return new(big.Int).Exp(c.params.G, x, c.params.N)
}

// A calculates A = g^a % N
func (c *SRPCore) A(a *big.Int) *big.Int {
	return new(big.Int).Exp(c.params.G, a, c.params.N)
}

// B calculates B = (k*v + g^b % N) % N
func (c *SRPCore) B(v, b *big.Int) (*big.Int, error) {
	k, err := c.K_multiplier()
	if err != nil {
		return nil, err
	}
	
	// k * v
	kv := new(big.Int).Mul(k, v)
	
	// g^b % N
	gb := new(big.Int).Exp(c.params.G, b, c.params.N)
	
	// (k*v + g^b) % N
	return new(big.Int).Add(kv, gb).Mod(new(big.Int).Add(kv, gb), c.params.N), nil
}

// U calculates u = H(A | B)
func (c *SRPCore) U(A, B *big.Int) (*big.Int, error) {
	return c.H(A, B)
}

// ClientS calculates S = (B - k*(g^x % N))^(a + u*x) % N
func (c *SRPCore) ClientS(B, x, a, u *big.Int) (*big.Int, error) {
	k, err := c.K_multiplier()
	if err != nil {
		return nil, err
	}
	
	// g^x % N
	gx := new(big.Int).Exp(c.params.G, x, c.params.N)
	
	// k * (g^x % N)
	kgx := new(big.Int).Mul(k, gx)
	
	// B - k*(g^x % N)
	base := new(big.Int).Sub(B, kgx)
	
	// a + u*x
	exp := new(big.Int).Add(a, new(big.Int).Mul(u, x))
	
	// (B - k*(g^x % N))^(a + u*x) % N
	return new(big.Int).Exp(base, exp, c.params.N), nil
}

// ServerS calculates S = (A * v^u % N)^b % N
func (c *SRPCore) ServerS(A, v, u, b *big.Int) *big.Int {
	// v^u % N
	vu := new(big.Int).Exp(v, u, c.params.N)
	
	// A * v^u % N
	base := new(big.Int).Mul(A, vu).Mod(new(big.Int).Mul(A, vu), c.params.N)
	
	// (A * v^u % N)^b % N
	return new(big.Int).Exp(base, b, c.params.N)
}

// K_multiplier calculates multiplier k = H(N | g)
func (c *SRPCore) K_multiplier() (*big.Int, error) {
	return c.H(c.params.N, c.params.G)
}

// K calculates shared key K = H(S)
func (c *SRPCore) K(S *big.Int) (*big.Int, error) {
	return c.H(S)
}

// M1 calculates first verification value M1 = H(A | B | K)
func (c *SRPCore) M1(A, B, K *big.Int) (*big.Int, error) {
	return c.H(A, B, K)
}

// M2 calculates second verification value M2 = H(A | M1 | K)
func (c *SRPCore) M2(A, M1, K *big.Int) (*big.Int, error) {
	return c.H(A, M1, K)
}

// IsZeroWhenModN checks if value is zero mod N
func (c *SRPCore) IsZeroWhenModN(n *big.Int) bool {
	return new(big.Int).Mod(n, c.params.N).Cmp(big.NewInt(0)) == 0
}

// getSRPParams gets SRP parameters
func getSRPParams(length SRPGroupLength) (*SRPParams, error) {
	switch length {
	case SRPGroup3072:
		return &SRPParams{
			Length: 3072,
			Hash:   "SHA-256",
			G:      big.NewInt(5),
			N:      hexToBigInt("FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1 29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245 E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE45B3D C2007CB8 A163BF05 98DA4836 1C55D39A 69163FA8 FD24CF5F 83655D23 DCA3AD96 1C62F356 208552BB 9ED52907 7096966D 670C354E 4ABC9804 F1746C08 CA18217C 32905E46 2E36CE3B E39E772C 180E8603 9B2783A2 EC07A28F B5C55DF0 6F4C52C9 DE2BCBF6 95581718 3995497C EA956AE5 15D22618 98FA0510 15728E5A 8AAAC42D AD33170D 04507A33 A85521AB DF1CBA64 ECFB8504 58DBEF0A 8AEA7157 5D060C7D B3970F85 A6E1E4C7 ABF5AE8C DB0933D7 1E8C94E0 4A25619D CEE3D226 1AD2EE6B F12FFA06 D98A0864 D8760273 3EC86A64 521F2B18 177B200C BBE11757 7A615D6C 770988C0 BAD946E2 08E24FA0 74E5AB31 43DB5BFC E0FD108E 4B82D120 A93AD2CA FFFFFFFF FFFFFFFF"),
		}, nil
		
	case SRPGroup4096:
		return &SRPParams{
			Length: 4096,
			Hash:   "SHA-256",
			G:      big.NewInt(5),
			N:      hexToBigInt("FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1 29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245 E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE45B3D C2007CB8 A163BF05 98DA4836 1C55D39A 69163FA8 FD24CF5F 83655D23 DCA3AD96 1C62F356 208552BB 9ED52907 7096966D 670C354E 4ABC9804 F1746C08 CA18217C 32905E46 2E36CE3B E39E772C 180E8603 9B2783A2 EC07A28F B5C55DF0 6F4C52C9 DE2BCBF6 95581718 3995497C EA956AE5 15D22618 98FA0510 15728E5A 8AAAC42D AD33170D 04507A33 A85521AB DF1CBA64 ECFB8504 58DBEF0A 8AEA7157 5D060C7D B3970F85 A6E1E4C7 ABF5AE8C DB0933D7 1E8C94E0 4A25619D CEE3D226 1AD2EE6B F12FFA06 D98A0864 D8760273 3EC86A64 521F2B18 177B200C BBE11757 7A615D6C 770988C0 BAD946E2 08E24FA0 74E5AB31 43DB5BFC E0FD108E 4B82D120 A9210801 1A723C12 A787E6D7 88719A10 BDBA5B26 99C32718 6AF4E23C 1A946834 B6150BDA 2583E9CA 2AD44CE8 DBBBC2DB 04DE8EF9 2E8EFC14 1FBECAA6 287C5947 4E6BC05D 99B2964F A090C3A2 233BA186 515BE7ED 1F612970 CEE2D7AF B81BDD76 2170481C D0069127 D5B05AA9 93B4EA98 8D8FDDC1 86FFB7DC 90A6C08F 4DF435C9 34063199 FFFFFFFF FFFFFFFF"),
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported group length: %d", length)
	}
}

// hexToBigInt helper function: converts hexadecimal string to big integer
func hexToBigInt(hex string) *big.Int {
	// Remove spaces
	cleanHex := ""
	for _, char := range hex {
		if char != ' ' {
			cleanHex += string(char)
		}
	}
	
	result, success := new(big.Int).SetString(cleanHex, 16)
	if !success {
		panic(fmt.Sprintf("Failed to parse hex string: %s", hex))
	}
	
	return result
}

// generateRandomBigInt generates random big integer
func generateRandomBigInt(bytes int) (*big.Int, error) {
	randomBytes := make([]byte, bytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	
	return new(big.Int).SetBytes(randomBytes), nil
}

// generateUUID generates UUID (simplified version)
func generateUUID() string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		randomBytes[0:4],
		randomBytes[4:6],
		randomBytes[6:8],
		randomBytes[8:10],
		randomBytes[10:16])
}

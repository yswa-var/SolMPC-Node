### **Networking & Communication Plan for Validators**
The goal is to ensure **secure and efficient communication** between validators and the aggregator while sharing **partial signatures**. We'll use **gRPC** for ease of setup, performance, and structured communication.

---

## **Step 1: Choosing the Communication Protocol**
We have two options:

1️⃣ **gRPC (Recommended)**
- High-performance, binary serialization (Protocol Buffers).
- Secure, supports authentication & streaming.
- Well-integrated in Go (`google.golang.org/grpc`).

2️⃣ **libp2p (Decentralized, Advanced)**
- Fully decentralized, ideal for **peer-to-peer (P2P)** communication.
- Harder to set up compared to gRPC.
- Useful for permissionless or highly resilient networks.

🔹 **For simplicity and ease of deployment, we will use gRPC.**

---

## **Step 2: gRPC-based Communication Flow**
### **Roles**
1. **Validators**
    - Sign transactions and send partial signatures to the aggregator.
    - Can also receive aggregated signatures from the aggregator.

2. **Aggregator**
    - Collects `t` valid partial signatures.
    - Aggregates and broadcasts the final signature to Solana.

---

## **Step 3: Define gRPC Service (Proto File)**
We'll define a **gRPC service** where validators send partial signatures.

📄 **`proto/validator.proto`**
```proto
syntax = "proto3";

package validator;

service ValidatorService {
  // Validator sends a partial signature
  rpc SendPartialSignature (SignatureRequest) returns (SignatureResponse);

  // Aggregator broadcasts the final aggregated signature
  rpc BroadcastFinalSignature (FinalSignatureRequest) returns (FinalSignatureResponse);
}

// Request to send a partial signature
message SignatureRequest {
  string validator_id = 1;
  bytes message = 2;
  bytes partial_signature = 3;
}

// Response after sending a signature
message SignatureResponse {
  bool success = 1;
  string message = 2;
}

// Request to broadcast final signature
message FinalSignatureRequest {
  bytes aggregated_signature = 1;
}

// Response after final signature broadcast
message FinalSignatureResponse {
  bool success = 1;
  string message = 2;
}
```
---

## **Step 4: Implement the Validator gRPC Client**
Validators will:  
✅ Sign transactions.  
✅ Send partial signatures to the aggregator via gRPC.

📄 **`validator_client.go`**
```go
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	pb "path/to/generated/proto"

	"google.golang.org/grpc"
)

// gRPC Server address (Aggregator)
const serverAddr = "localhost:50051"

// Simulate signing
func signMessage(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

func main() {
	// Connect to the Aggregator
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewValidatorServiceClient(conn)

	// Generate key pair (for demo purposes)
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	message := []byte("Transaction data")
	partialSignature := signMessage(privateKey, message)

	req := &pb.SignatureRequest{
		ValidatorId:      "validator_1",
		Message:          message,
		PartialSignature: partialSignature,
	}

	// Send partial signature
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.SendPartialSignature(ctx, req)
	if err != nil {
		log.Fatalf("Error sending signature: %v", err)
	}

	fmt.Println("Response from Aggregator:", resp.Message)
}
```
---

## **Step 5: Implement the Aggregator gRPC Server**
The **Aggregator** will:  
✅ Receive partial signatures.  
✅ Validate and store signatures.  
✅ Aggregate and broadcast the final signature.

📄 **`aggregator_server.go`**
```go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "path/to/generated/proto"

	"google.golang.org/grpc"
)

// Aggregator struct
type AggregatorServer struct {
	pb.UnimplementedValidatorServiceServer
	mu               sync.Mutex
	partialSignatures map[string][]byte
}

// Receive a partial signature from a validator
func (s *AggregatorServer) SendPartialSignature(ctx context.Context, req *pb.SignatureRequest) (*pb.SignatureResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.partialSignatures[req.ValidatorId] = req.PartialSignature

	fmt.Println("Received partial signature from:", req.ValidatorId)

	return &pb.SignatureResponse{Success: true, Message: "Signature received"}, nil
}

// Start gRPC server
func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	aggregator := &AggregatorServer{partialSignatures: make(map[string][]byte)}

	pb.RegisterValidatorServiceServer(server, aggregator)

	fmt.Println("Aggregator is running on port 50051...")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
```
---

## **Step 6: Broadcasting the Final Signature to Solana**
Once enough (`t`) partial signatures are received, the aggregator **aggregates the final signature** and broadcasts it to Solana.

🔹 Use Solana’s JSON-RPC API:
```go
import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Send final signature to Solana
func BroadcastFinalSignature(signature []byte) {
	url := "https://api.mainnet-beta.solana.com" // Update for your network

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "sendTransaction",
		"params":  []interface{}{signature},
	}

	body, _ := json.Marshal(payload)
	resp, _ := http.Post(url, "application/json", bytes.NewBuffer(body))

	defer resp.Body.Close()
}
```

---

## **Step 7: Testing**
1. **Run the Aggregator**
   ```sh
   go run aggregator_server.go
   ```
2. **Run Multiple Validators**
   ```sh
   go run validator_client.go
   ```
3. **Check Logs** for received partial signatures.
4. **Verify Final Signature Aggregation.**

---

## **Final Checklist ✅**
✔ Define `validator.proto` for gRPC.  
✔ Implement **Validator gRPC Client**.  
✔ Implement **Aggregator gRPC Server**.  
✔ Send partial signatures via gRPC.  
✔ Aggregate and **broadcast final signature** to Solana.  
✔ Test with multiple validators.

🚀 **Next Steps?**
- Do you want me to refine the **Solana broadcasting mechanism** further?
- Need help with **libp2p setup** if you prefer P2P?

Let me know! 🎯

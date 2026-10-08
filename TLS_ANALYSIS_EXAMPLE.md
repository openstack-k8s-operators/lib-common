# TLS Analysis Feature - Usage Examples

This document demonstrates how to use the TLS 1.3 detection and PQC-safe algorithm checking features.

## Overview

The TLS library now provides functionality to:
- Detect if certificates support TLS 1.3
- Check if certificates use Post-Quantum Cryptography (PQC) safe algorithms
- Analyze certificate properties (key algorithm, key size, signature algorithm)
- Get recommended cipher suites based on certificate strength

## Usage Examples

### 1. Basic Certificate Analysis

```go
import (
    "context"
    "fmt"
    
    "github.com/openstack-k8s-operators/lib-common/modules/common/tls"
    "k8s.io/apimachinery/pkg/types"
)

// Analyze a certificate from PEM bytes
func analyzeCertFromPEM(certPEM []byte) error {
    analysis, err := tls.AnalyzeCertificate(certPEM)
    if err != nil {
        return fmt.Errorf("failed to analyze certificate: %w", err)
    }
    
    fmt.Printf("TLS 1.3 Support: %v\n", analysis.SupportsTLS13)
    fmt.Printf("PQC-Safe: %v\n", analysis.IsPQCSafe)
    fmt.Printf("Key Algorithm: %s\n", analysis.KeyAlgorithm)
    fmt.Printf("Key Size: %d bits\n", analysis.KeySize)
    fmt.Printf("Signature Algorithm: %s\n", analysis.SignatureAlgorithm)
    fmt.Printf("Recommended Cipher Suites: %v\n", analysis.CipherSuites)
    
    return nil
}
```

### 2. Analyze Certificate from Kubernetes Secret

```go
// Analyze a certificate stored in a Kubernetes secret
func analyzeCertFromSecret(ctx context.Context, client client.Client, namespace, secretName string) error {
    analysis, err := tls.AnalyzeCertSecret(
        ctx,
        client,
        types.NamespacedName{Name: secretName, Namespace: namespace},
    )
    if err != nil {
        return fmt.Errorf("failed to analyze certificate secret: %w", err)
    }
    
    if !analysis.SupportsTLS13 {
        fmt.Printf("Warning: Certificate does not support TLS 1.3\n")
    }
    
    if !analysis.IsPQCSafe {
        fmt.Printf("Warning: Certificate is not PQC-safe. Consider using:\n")
        fmt.Printf("  - RSA with >= 3072 bits\n")
        fmt.Printf("  - ECDSA with P-384 or P-521 curves\n")
    }
    
    return nil
}
```

### 3. Using Service Methods in an Operator

```go
import (
    "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
    "github.com/openstack-k8s-operators/lib-common/modules/common/tls"
    ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MyServiceReconciler) reconcileTLS(
    ctx context.Context,
    h *helper.Helper,
    instance *myv1.MyService,
) (ctrl.Result, error) {
    
    tlsService := &tls.Service{
        SecretName: instance.Spec.TLS.SecretName,
    }
    
    // Check if TLS 1.3 is enabled
    isTLS13, err := tlsService.IsTLS13Enabled(ctx, h, instance.Namespace)
    if err != nil {
        return ctrl.Result{}, err
    }
    
    // Check if using PQC-safe algorithms
    isPQC, err := tlsService.IsPQCSafe(ctx, h, instance.Namespace)
    if err != nil {
        return ctrl.Result{}, err
    }
    
    // Get full analysis
    analysis, err := tlsService.GetTLSAnalysis(ctx, h, instance.Namespace)
    if err != nil {
        return ctrl.Result{}, err
    }
    
    // Log the findings
    r.Log.Info("TLS Configuration Analysis",
        "service", instance.Name,
        "tls13Enabled", isTLS13,
        "pqcSafe", isPQC,
        "keyAlgorithm", analysis.KeyAlgorithm,
        "keySize", analysis.KeySize,
    )
    
    // Update status or conditions based on findings
    if !isPQC {
        // Set a warning condition or status
        r.Log.Info("Certificate is not PQC-safe",
            "recommendation", "Consider upgrading to RSA-3072+ or ECDSA P-384+")
    }
    
    return ctrl.Result{}, nil
}
```

### 4. Validation During Certificate Creation

```go
func validateCertificateRequirements(ctx context.Context, client client.Client, certSecret types.NamespacedName, requirePQC bool) error {
    analysis, err := tls.AnalyzeCertSecret(ctx, client, certSecret)
    if err != nil {
        return fmt.Errorf("failed to analyze certificate: %w", err)
    }
    
    if !analysis.SupportsTLS13 {
        return fmt.Errorf("certificate does not support TLS 1.3")
    }
    
    if requirePQC && !analysis.IsPQCSafe {
        return fmt.Errorf("certificate is not PQC-safe (current: %s-%d, required: RSA-3072+ or ECDSA P-384+)",
            analysis.KeyAlgorithm, analysis.KeySize)
    }
    
    return nil
}
```

## PQC-Safe Algorithm Requirements

Based on NIST SP 800-57 guidelines for transitional quantum resistance:

### RSA
- **Not PQC-Safe**: RSA-1024, RSA-2048
- **PQC-Safe**: RSA-3072, RSA-4096

### ECDSA
- **Not PQC-Safe**: P-256
- **PQC-Safe**: P-384, P-521

### Ed25519
- **Not PQC-Safe**: Ed25519 (256-bit)

### TLS 1.3 Compatibility

All modern key algorithms are compatible with TLS 1.3:
- RSA >= 2048 bits
- ECDSA with P-256, P-384, or P-521
- Ed25519

## Testing

The implementation includes comprehensive unit tests covering:
- Certificate analysis for all key types (RSA, ECDSA, Ed25519)
- Different key sizes
- PQC-safety validation
- TLS 1.3 compatibility checks
- Error handling for invalid certificates
- Cipher suite recommendations

Run tests with:
```bash
cd modules/common
go test -v ./tls -run TestAnalyzeCertificate
go test -v ./tls -run TestPQCSafe
go test -v ./tls -run TestTLS13Compatibility
```

## Certificate Generation for Testing

Use the test helpers to generate certificates:

```go
import (
    helpers "github.com/openstack-k8s-operators/lib-common/modules/common/test/helpers"
)

// Generate RSA 3072-bit certificate (PQC-safe)
cert, err := helpers.GenerateCertificate(helpers.RSA3072CertConfig())

// Generate ECDSA P-384 certificate (PQC-safe)
cert, err := helpers.GenerateCertificate(helpers.ECDSAP384CertConfig())

// Generate custom certificate
config := &helpers.CertConfig{
    KeyType:      "rsa",
    KeySize:      4096,
    CommonName:   "my-service.example.com",
    DNSNames:     []string{"my-service.example.com", "localhost"},
    Organization: "My Organization",
    NotBefore:    time.Now(),
    NotAfter:     time.Now().Add(365 * 24 * time.Hour),
}
cert, err := helpers.GenerateCertificate(config)
```

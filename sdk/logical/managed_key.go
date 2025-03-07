package logical

import (
	"context"
	"crypto"
	"io"
)

type KeyUsage int

const (
	KeyUsageEncrypt KeyUsage = 1 + iota
	KeyUsageDecrypt
	KeyUsageSign
	KeyUsageVerify
	KeyUsageWrap
	KeyUsageUnwrap
)

type ManagedKey interface {
	// Name is a human-readable identifier for a managed key that may change/renamed. Use Uuid if a
	// long term consistent identifier is needed.
	Name() string
	// UUID is a unique identifier for a managed key that is guaranteed to remain
	// consistent even if a key is migrated or renamed.
	UUID() string
	// Present returns true if the key is established in the KMS.  This may return false if for example
	// an HSM library is not configured on all cluster nodes.
	Present(ctx context.Context) (bool, error)
	Finalize(context.Context) error

	// AllowsAll returns true if all the requested usages are supported by the managed key.
	AllowsAll(usages []KeyUsage) bool
}

type ManagedKeySystemView interface {
	GetManagedKeyByName(ctx context.Context, keyName, mountPoint string) (ManagedKey, error)
	GetManagedKeyByUUID(ctx context.Context, keyUuid, mountPoint string) (ManagedKey, error)
}

type ManagedAsymmetricKey interface {
	ManagedKey
	GetPublicKey(ctx context.Context) (crypto.PublicKey, error)
}

type ManagedKeyLifecycle interface {
	// GenerateKey generates a key in the KMS if it didn't yet exist, returning the id.
	// If it already existed, returns the existing id.  KMSKey's key material is ignored if present.
	GenerateKey(ctx context.Context) (string, error)
}

type ManagedSigningKey interface {
	ManagedAsymmetricKey

	// Sign returns a digital signature of the provided value.  The SignerOpts param must provide the hash function
	// that generated the value (if any).
	// The optional randomSource specifies the source of random values and may be ignored by the implementation
	// (such as on HSMs with their own internal RNG)
	Sign(ctx context.Context, value []byte, randomSource io.Reader, opts crypto.SignerOpts) ([]byte, error)

	// Verify verifies the provided signature against the value.  The SignerOpts param must provide the hash function
	// that generated the value (if any).
	// If true is returned the signature is correct, false otherwise.
	Verify(ctx context.Context, signature, value []byte, opts crypto.SignerOpts) (bool, error)

	// GetSigner returns an implementation of crypto.Signer backed by the managed key.  This should be called
	// as needed so as to use per request contexts.
	GetSigner(context.Context) (crypto.Signer, error)
}

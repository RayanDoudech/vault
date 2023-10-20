// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package vault

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
)

// SealConfig is used to describe the seal configuration
type SealConfig struct {
	// The type, for sanity checking. See SealConfigType for valid values.
	Type string `json:"type" mapstructure:"type"`

	// SecretShares is the number of shares the secret is split into. This is
	// the N value of Shamir.
	SecretShares int `json:"secret_shares" mapstructure:"secret_shares"`

	// SecretThreshold is the number of parts required to open the vault. This
	// is the T value of Shamir.
	SecretThreshold int `json:"secret_threshold" mapstructure:"secret_threshold"`

	// PGPKeys is the array of public PGP keys used, if requested, to encrypt
	// the output unseal tokens. If provided, it sets the value of
	// SecretShares. Ordering is important.
	PGPKeys []string `json:"pgp_keys" mapstructure:"pgp_keys"`

	// Nonce is a nonce generated by Vault used to ensure that when unseal keys
	// are submitted for a rekey operation, the rekey operation itself is the
	// one intended. This prevents hijacking of the rekey operation, since it
	// is unauthenticated.
	Nonce string `json:"nonce" mapstructure:"nonce"`

	// Backup indicates whether or not a backup of PGP-encrypted unseal keys
	// should be stored at coreUnsealKeysBackupPath after successful rekeying.
	Backup bool `json:"backup" mapstructure:"backup"`

	// How many keys to store, for seals that support storage.  Always 0 or 1.
	StoredShares int `json:"stored_shares" mapstructure:"stored_shares"`

	// Stores the progress of the rekey operation (key shares)
	RekeyProgress [][]byte `json:"-"`

	// VerificationRequired indicates that after a rekey validation must be
	// performed (via providing shares from the new key) before the new key is
	// actually installed. This is omitted from JSON as we don't persist the
	// new key, it lives only in memory.
	VerificationRequired bool `json:"-"`

	// VerificationKey is the new key that we will roll to after successful
	// validation
	VerificationKey []byte `json:"-"`

	// VerificationNonce stores the current operation nonce for verification
	VerificationNonce string `json:"-"`

	// Stores the progress of the verification operation (key shares)
	VerificationProgress [][]byte `json:"-"`

	// Name is the name provided in the seal configuration to identify the seal
	Name string `json:"name" mapstructure:"name"`
}

// Validate is used to sanity check the seal configuration
func (s *SealConfig) Validate() error {
	if s.SecretShares < 1 {
		return fmt.Errorf("shares must be at least one")
	}
	if s.SecretThreshold < 1 {
		return fmt.Errorf("threshold must be at least one")
	}
	if s.SecretShares > 1 && s.SecretThreshold == 1 {
		return fmt.Errorf("threshold must be greater than one for multiple shares")
	}
	if s.SecretShares > 255 {
		return fmt.Errorf("shares must be less than 256")
	}
	if s.SecretThreshold > 255 {
		return fmt.Errorf("threshold must be less than 256")
	}
	if s.SecretThreshold > s.SecretShares {
		return fmt.Errorf("threshold cannot be larger than shares")
	}
	if s.StoredShares > 1 {
		return fmt.Errorf("stored keys cannot be larger than 1")
	}
	if len(s.PGPKeys) > 0 && len(s.PGPKeys) != s.SecretShares {
		return fmt.Errorf("count mismatch between number of provided PGP keys and number of shares")
	}
	if len(s.PGPKeys) > 0 {
		for _, keystring := range s.PGPKeys {
			data, err := base64.StdEncoding.DecodeString(keystring)
			if err != nil {
				return fmt.Errorf("error decoding given PGP key: %w", err)
			}
			_, err = openpgp.ReadEntity(packet.NewReader(bytes.NewBuffer(data)))
			if err != nil {
				return fmt.Errorf("error parsing given PGP key: %w", err)
			}
		}
	}
	return nil
}

func (s *SealConfig) Clone() *SealConfig {
	ret := &SealConfig{
		Type:                 s.Type,
		SecretShares:         s.SecretShares,
		SecretThreshold:      s.SecretThreshold,
		Nonce:                s.Nonce,
		Backup:               s.Backup,
		StoredShares:         s.StoredShares,
		VerificationRequired: s.VerificationRequired,
		VerificationNonce:    s.VerificationNonce,
		Name:                 s.Name,
	}
	if len(s.PGPKeys) > 0 {
		ret.PGPKeys = make([]string, len(s.PGPKeys))
		copy(ret.PGPKeys, s.PGPKeys)
	}
	if len(s.VerificationKey) > 0 {
		ret.VerificationKey = make([]byte, len(s.VerificationKey))
		copy(ret.VerificationKey, s.VerificationKey)
	}
	return ret
}

// SealConfigType specifies the "type" of a seal according to the following rules:
// - For a defaultSeal, the type is SealConfigTypeShamir, since all defaultSeals use a shamir wrapper.
//
// - For an autoseal:
//   - if there is a single encryption wrapper, the type is the wrapper type
//   - if there are two or more encryption wrappers, the type is SealConfigTypeMultiseal
//
// - For a recovery seal, the type is SealConfigTypeShamir, since all recovery seals are defaultSeals.
type SealConfigType string

const (
	SealConfigTypeMultiseal         = SealConfigType("multiseal")
	SealConfigTypeShamir            = SealConfigType(wrapping.WrapperTypeShamir)
	SealConfigTypePkcs11            = SealConfigType(wrapping.WrapperTypePkcs11)
	SealConfigTypeAwsKms            = SealConfigType(wrapping.WrapperTypeAwsKms)
	SealConfigTypeHsmAutoDeprecated = SealConfigType(wrapping.WrapperTypeHsmAuto)
	SealConfigTypeTransit           = SealConfigType(wrapping.WrapperTypeTransit)
	SealConfigTypeGcpCkms           = SealConfigType(wrapping.WrapperTypeGcpCkms)

	// SealConfigTypeRecovery is an alias for SealConfigTypeShamir since all recovery seals are
	// defaultSeals using shamir wrappers.
	SealConfigTypeRecovery = SealConfigTypeShamir

	// SealConfigTypeRecoveryUnsupported is for convenience.
	SealConfigTypeRecoveryUnsupported = SealConfigType("unsupported")
)

func (s SealConfigType) String() string {
	return string(s)
}

func (s SealConfigType) IsSameAs(t string) bool {
	return s.String() == t
}

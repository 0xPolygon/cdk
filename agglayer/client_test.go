package agglayer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultNilFieldsInCertificate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		certificate *SignedCertificate
		expectedErr error
	}{
		{
			name:        "Nil certificate",
			certificate: nil,
			expectedErr: errors.New("certificate is nil"),
		},
		{
			name: "Nil inner certificate",
			certificate: &SignedCertificate{
				Certificate: nil,
			},
			expectedErr: errors.New("certificate is nil"),
		},
		{
			name: "Nil signature",
			certificate: &SignedCertificate{
				Certificate: &Certificate{},
				Signature:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Nil BridgeExits",
			certificate: &SignedCertificate{
				Certificate: &Certificate{
					BridgeExits: nil,
				},
				Signature: &Signature{},
			},
			expectedErr: nil,
		},
		{
			name: "Nil ImportedBridgeExits",
			certificate: &SignedCertificate{
				Certificate: &Certificate{
					ImportedBridgeExits: nil,
				},
				Signature: &Signature{},
			},
			expectedErr: nil,
		},
		{
			name: "All fields non-nil",
			certificate: &SignedCertificate{
				Certificate: &Certificate{
					BridgeExits:         []*BridgeExit{},
					ImportedBridgeExits: []*ImportedBridgeExit{},
				},
				Signature: &Signature{},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := defaultNilFieldsInCertificate(tt.certificate)
			assert.Equal(t, tt.expectedErr, err)

			if tt.certificate != nil && tt.certificate.Certificate != nil {
				assert.NotNil(t, tt.certificate.Signature)
				assert.NotNil(t, tt.certificate.Certificate.BridgeExits)
				assert.NotNil(t, tt.certificate.Certificate.ImportedBridgeExits)
			}
		})
	}
}

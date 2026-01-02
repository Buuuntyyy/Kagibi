package utils

import (
	"testing"
)

func TestSecureJoin(t *testing.T) {
	tests := []struct {
		name       string
		root       string
		unsafePath string
		wantErr    bool
	}{
		{
			name:       "Chemin valide simple",
			root:       "/var/www",
			unsafePath: "image.png",
			wantErr:    false,
		},
		{
			name:       "Chemin valide sous-dossier",
			root:       "/var/www",
			unsafePath: "uploads/image.png",
			wantErr:    false,
		},
		{
			name:       "Attaque Path Traversal simple",
			root:       "/var/www",
			unsafePath: "../etc/passwd",
			wantErr:    true,
		},
		{
			name:       "Attaque Path Traversal complexe",
			root:       "/var/www",
			unsafePath: "uploads/../../etc/passwd",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SecureJoin(tt.root, tt.unsafePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureJoin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

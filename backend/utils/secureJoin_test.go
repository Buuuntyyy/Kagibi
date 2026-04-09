package utils

import (
	"testing"
)

const testRootPath = "/var/www"

func TestSecureJoin(t *testing.T) {
	tests := []struct {
		name       string
		root       string
		unsafePath string
		wantErr    bool
	}{
		{
			name:       "Chemin valide simple",
			root:       testRootPath,
			unsafePath: "image.png",
			wantErr:    false,
		},
		{
			name:       "Chemin valide sous-dossier",
			root:       testRootPath,
			unsafePath: "uploads/image.png",
			wantErr:    false,
		},
		{
			name:       "Attaque Path Traversal simple",
			root:       testRootPath,
			unsafePath: "../etc/passwd",
			wantErr:    true,
		},
		{
			name:       "Attaque Path Traversal complexe",
			root:       testRootPath,
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

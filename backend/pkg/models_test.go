package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	t.Run("User Creation", func(t *testing.T) {
		user := User{
			Email:     "test@example.com",
			PublicKey: "public-key-123",
		}

		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "public-key-123", user.PublicKey)
	})
}

func TestFileModel(t *testing.T) {
	t.Run("File Creation", func(t *testing.T) {
		file := File{
			Name:         "test.txt",
			EncryptedKey: "encrypted-key-123",
			Size:         1024,
		}

		assert.Equal(t, "test.txt", file.Name)
		assert.Equal(t, "encrypted-key-123", file.EncryptedKey)
		assert.Equal(t, int64(1024), file.Size)
	})
}

func TestFolderModel(t *testing.T) {
	t.Run("Folder Creation", func(t *testing.T) {
		folder := Folder{
			Name: "Documents",
		}

		assert.Equal(t, "Documents", folder.Name)
	})
}

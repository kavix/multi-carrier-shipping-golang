package utils

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func GenerateID() string {
	return uuid.New().String()
}

func RandomDigits(n int) string {
	digits := "0123456789"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		result[i] = digits[num.Int64()]
	}
	return string(result)
}

func RandomAlphanumeric(n int) string {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}
	return string(result)
}

func InitDB(db *sql.DB, migrationsDir string) error {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				return fmt.Errorf("read migration file %s: %w", file.Name(), err)
			}

			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("execute migration %s: %w", file.Name(), err)
			}
		}
	}
	return nil
}

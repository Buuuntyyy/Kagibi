// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package files

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

const ENDPOINT = "/rename"

func TestRenameHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup Mock DB
	sqldb, mockDB, _ := sqlmock.New()
	defer sqldb.Close()
	db := bun.NewDB(sqldb, pgdialect.New())

	// Setup Mock Redis
	redisClient, _ := redismock.NewClientMock()

	r := gin.New()
	r.POST(ENDPOINT, func(c *gin.Context) {
		c.Set("user_id", "user-123")
		RenameHandler(c, db, redisClient)
	})

	t.Run("Rename File Success", func(t *testing.T) {
		reqBody := RenameRequest{
			ID:      1,
			Type:    "file",
			NewName: "NewName.txt",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", ENDPOINT, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// 1. Expect GetFile
		// Bun génère souvent des requêtes avec des alias (f.id, f.name...)
		mockDB.ExpectQuery(`SELECT .* FROM "files"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "user_id"}).
				AddRow(1, "OldName.txt", "/root/OldName.txt", "user-123"))

		// 2. Check for name conflicts
		mockDB.ExpectQuery(`SELECT .* FROM "files"`).
			WillReturnRows(sqlmock.NewRows([]string{})) // No conflict

		// 3. Begin Transaction
		mockDB.ExpectBegin()

		// 4. Expect Update File
		mockDB.ExpectExec(`UPDATE "files"`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 5. Commit Transaction
		mockDB.ExpectCommit()

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid Name", func(t *testing.T) {
		reqBody := RenameRequest{
			ID:      1,
			Type:    "file",
			NewName: "Bad/Name",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", ENDPOINT, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

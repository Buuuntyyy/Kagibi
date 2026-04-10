// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package folders

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

const foldersPath = "/folders"

func TestCreateHandler(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)

	// Setup Mock DB
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqldb.Close()
	db := bun.NewDB(sqldb, pgdialect.New())

	// Setup Router
	r := gin.New()
	r.POST(foldersPath, func(c *gin.Context) {
		// Mock Auth Middleware
		c.Set("user_id", "user-123")
		CreateHandler(c, db)
	})

	t.Run("Success", func(t *testing.T) {
		// Prepare Request
		reqBody := CreateFolderRequest{
			Name: "NewFolder",
			Path: "/root",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", foldersPath, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Mock DB Expectations
		// Expect INSERT into folders and RETURNING id
		mock.ExpectQuery(`INSERT INTO "folders"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect INSERT into folder_sizes
		mock.ExpectQuery(`INSERT INTO "folder_sizes"`).
			WillReturnRows(sqlmock.NewRows([]string{"folder_id"}).AddRow(1))

		// Execute
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		// Cleanup: Remove created folder on disk
		os.RemoveAll("uploads/user-123/root/NewFolder")
	})

	t.Run("Invalid Name (XSS)", func(t *testing.T) {
		reqBody := CreateFolderRequest{
			Name: "<script>alert(1)</script>",
			Path: "/root",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", foldersPath, bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

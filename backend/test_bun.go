package main

import (
"database/sql"
"fmt"
"kagibi/backend/handlers/users"
"github.com/uptrace/bun"
"github.com/uptrace/bun/dialect/pgdialect"
)

func main() {
db := bun.NewDB(sql.OpenDB(nil), pgdialect.New())
settings := users.SecuritySettings{UserID: "123"}

q := db.NewInsert().
Model(&settings).
On("CONFLICT (user_id) DO UPDATE").
Set("mfa_enabled = EXCLUDED.mfa_enabled").
Set("mfa_verified = EXCLUDED.mfa_verified").
Set("require_mfa_on_login = EXCLUDED.require_mfa_on_login").
Set("require_mfa_on_destructive_actions = EXCLUDED.require_mfa_on_destructive_actions").
Set("require_mfa_on_downloads = EXCLUDED.require_mfa_on_downloads")

str := q.String()
fmt.Println(str)
}

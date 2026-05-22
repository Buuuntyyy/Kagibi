// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// admin is the operator CLI for Kagibi.
// Run it directly on the server (via SSH) — it connects to the same database
// as the backend and never exposes an HTTP surface.
//
// Usage:
//
//	./admin org create  --name <name> [--desc <description>] [--quota <mb>] [--owner-email <email>]
//	./admin org list
//	./admin org quota   --id <org_id> --quota <mb>
//	./admin org delete  --id <org_id>
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"kagibi/backend/internal/provisioning"
	"kagibi/backend/pkg"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
)

func main() {
	// Load .env if present (local dev). Production uses real env vars.
	_ = godotenv.Load()

	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	resource := os.Args[1] // "org"
	action := os.Args[2]   // "create" | "list" | "quota" | "delete"
	args := os.Args[3:]

	switch resource {
	case "org":
		runOrg(action, args)
	default:
		fmt.Fprintf(os.Stderr, "unknown resource %q\n", resource)
		printUsage()
		os.Exit(1)
	}
}

func runOrg(action string, args []string) {
	switch action {
	case "create":
		orgCreate(args)
	case "list":
		orgList()
	case "quota":
		orgQuota(args)
	case "delete":
		orgDelete(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown org action %q\n", action)
		printOrgUsage()
		os.Exit(1)
	}
}

// ── org create ────────────────────────────────────────────────────────────────

func orgCreate(args []string) {
	fs := flag.NewFlagSet("org create", flag.ExitOnError)
	name := fs.String("name", "", "organisation name (required)")
	desc := fs.String("desc", "", "organisation description")
	quota := fs.Int64("quota", 10240, "storage quota in MB (default: 10240 = 10 GB)")
	email := fs.String("owner-email", "", "owner email — receives the invitation link")
	_ = fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required")
		fs.Usage()
		os.Exit(1)
	}

	db := connectDB()
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := provisioning.CreateOrg(ctx, db, *name, *desc, *quota, *email)
	if err != nil {
		fatalf("create org: %v", err)
	}

	fmt.Println()
	fmt.Printf("  ✓ Organisation créée\n")
	fmt.Printf("  %-16s %d\n", "ID:", result.Org.ID)
	fmt.Printf("  %-16s %s\n", "Nom:", result.Org.Name)
	fmt.Printf("  %-16s %d Mo\n", "Quota:", result.Org.StorageQuotaMB)
	fmt.Println()
	fmt.Printf("  Lien d'invitation owner (valable 7 jours, 1 usage) :\n")
	fmt.Printf("  %s\n", result.InviteURL)
	if *email != "" {
		fmt.Printf("\n  Un email a été envoyé à %s\n", *email)
	}
	fmt.Println()
}

// ── org list ──────────────────────────────────────────────────────────────────

func orgList() {
	db := connectDB()
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	orgs, err := provisioning.ListOrgs(ctx, db)
	if err != nil {
		fatalf("list orgs: %v", err)
	}

	if len(orgs) == 0 {
		fmt.Println("Aucune organisation.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNOM\tMEMBRES\tUTILISÉ (Mo)\tQUOTA (Mo)\tCRÉÉ LE")
	for _, o := range orgs {
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%d\t%s\n",
			o.ID, o.Name, o.MemberCount,
			o.StorageUsedMB, o.StorageQuotaMB,
			o.CreatedAt.Format("2006-01-02"),
		)
	}
	_ = w.Flush()
}

// ── org quota ─────────────────────────────────────────────────────────────────

func orgQuota(args []string) {
	fs := flag.NewFlagSet("org quota", flag.ExitOnError)
	id := fs.Int64("id", 0, "organisation ID (required)")
	quota := fs.Int64("quota", 0, "new quota in MB (required)")
	_ = fs.Parse(args)

	if *id == 0 || *quota == 0 {
		fmt.Fprintln(os.Stderr, "error: --id and --quota are required")
		fs.Usage()
		os.Exit(1)
	}

	db := connectDB()
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := provisioning.SetOrgQuota(ctx, db, *id, *quota); err != nil {
		fatalf("set quota: %v", err)
	}

	fmt.Printf("✓ Quota de l'organisation %d mis à jour : %d Mo\n", *id, *quota)
}

// ── org delete ────────────────────────────────────────────────────────────────

func orgDelete(args []string) {
	fs := flag.NewFlagSet("org delete", flag.ExitOnError)
	id := fs.Int64("id", 0, "organisation ID (required)")
	yes := fs.Bool("yes", false, "skip confirmation prompt")
	_ = fs.Parse(args)

	if *id == 0 {
		fmt.Fprintln(os.Stderr, "error: --id is required")
		fs.Usage()
		os.Exit(1)
	}

	if !*yes {
		fmt.Printf("Supprimer définitivement l'organisation %d ? [oui/N] : ", *id)
		var ans string
		fmt.Scanln(&ans)
		if ans != "oui" {
			fmt.Println("Annulé.")
			return
		}
	}

	db := connectDB()
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := provisioning.DeleteOrg(ctx, db, *id); err != nil {
		fatalf("delete org: %v", err)
	}

	fmt.Printf("✓ Organisation %d supprimée.\n", *id)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func connectDB() *bun.DB {
	db := pkg.NewDB()
	// Quick ping to fail fast with a clear message.
	if err := db.Ping(); err != nil {
		fatalf("impossible de se connecter à la base de données : %v\n(vérifiez DATABASE_URL)", err)
	}
	return db
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "erreur: "+format+"\n", args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `admin — CLI opérateur Kagibi

Usage:
  admin <resource> <action> [flags]

Ressources:
  org   Gestion des organisations`)
	printOrgUsage()
}

func printOrgUsage() {
	fmt.Fprintln(os.Stderr, `
Actions org:
  create  --name <nom> [--desc <desc>] [--quota <Mo>] [--owner-email <email>]
  list
  quota   --id <id> --quota <Mo>
  delete  --id <id> [--yes]`)
}

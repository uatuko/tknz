package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"

	"go.tknz.dev/internal/db"
)

const (
	// Ref: https://www.rfc-editor.org/rfc/rfc9106.html#section-4 (second recommended option)
	argon2Time    = 3
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16

	dbTimeout = 5 * time.Second
)

func main() {
	idnId := flag.String("idn-id", "", "identity id")
	printJson := flag.Bool("print", false, "print password credentials as json to stdout instead of updating the db")
	flag.Parse()

	if !*printJson && *idnId == "" {
		fmt.Fprintf(os.Stderr, "[error] missing identity id\n")
		flag.Usage()
		os.Exit(1)
	}

	var src *db.IdnSrc
	if !*printJson {
		ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
		defer cancel()

		if err := db.Init(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "[error] failed to initialise db, err: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		s, err := db.RetrieveIdnSrcByIdnIdAndProviderSlug(ctx, *idnId, db.ProviderSlugPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] failed to retrieve password identity source, err: %v\n", err)
			os.Exit(1)
		}

		src = s
	}

	password := readPassword("New password: ")
	if password == "" {
		fmt.Fprintf(os.Stderr, "No password supplied\n")
		os.Exit(1)
	}

	if readPassword("Retype new password: ") != password {
		fmt.Fprintf(os.Stderr, "Sorry, passwords do not match.\n")
		os.Exit(1)
	}

	pwd := db.IdnSrcAttrsPwd{
		Typ:     db.IdnSrcAttrsPwdTypArgon2,
		Salt:    make([]byte, argon2SaltLen),
		Time:    argon2Time,
		Memory:  argon2Memory,
		Threads: argon2Threads,
	}

	rand.Read(pwd.Salt)
	pwd.Key = argon2.IDKey([]byte(password), pwd.Salt, pwd.Time, pwd.Memory, pwd.Threads, argon2KeyLen)

	if *printJson {
		b, err := json.Marshal(pwd)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s\n", b)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if err := src.SetAttrsPwd(ctx, pwd); err != nil {
		fmt.Fprintf(os.Stderr, "[error] failed to update password, err: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "passwd: password updated successfully\n")
}

func readPassword(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] failed to read password, err: %v\n", err)
		os.Exit(1)
	}

	return string(b)
}

// firestorecheck verifies read access without creating or modifying documents.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	document := flag.String("document", "", "document path to read; contents will not be printed")
	flag.Parse()
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if strings.TrimSpace(project) == "" {
		return errors.New("missing GOOGLE_CLOUD_PROJECT")
	}
	if flag.NArg() != 0 || !validDocumentPath(*document) {
		return errors.New("provide -document collection/document with no positional arguments")
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return errors.New("cannot initialize Firestore; check ADC and project configuration")
	}
	defer client.Close()
	_, err = client.Doc(*document).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return errors.New("Firestore returned NotFound; confirm database and document exist; read access not confirmed")
	}
	if err != nil {
		return errors.New("Firestore read failed; check project, database, IAM and connectivity")
	}
	fmt.Println("Firestore read access confirmed; document exists; contents not printed; no data written")
	return nil
}

func validDocumentPath(path string) bool {
	segments := strings.Split(path, "/")
	if len(segments) < 2 || len(segments)%2 != 0 {
		return false
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

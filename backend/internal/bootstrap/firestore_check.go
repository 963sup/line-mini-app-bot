// firestorecheck verifies read access without creating or modifying documents.
package bootstrap

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

// CheckFirestore confirms read access to one explicitly selected document.
func CheckFirestore(parent context.Context, args []string) (resultErr error) {
	flags := flag.NewFlagSet("check-firestore", flag.ContinueOnError)
	document := flags.String("document", "", "document path to read; contents will not be printed")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if strings.TrimSpace(project) == "" {
		return errors.New("missing GOOGLE_CLOUD_PROJECT")
	}
	if flags.NArg() != 0 || !validDocumentPath(*document) {
		return errors.New("provide -document collection/document with no positional arguments")
	}
	database := os.Getenv("FIRESTORE_DATABASE_ID")
	if database == "" {
		database = "(default)"
	}
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	client, err := firestore.NewClientWithDatabase(ctx, project, database)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := client.Close(); resultErr == nil {
			resultErr = closeErr
		}
	}()
	_, err = client.Doc(*document).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return errors.New("firestore returned NotFound; confirm database and document exist; read access not confirmed")
	}
	if err != nil {
		return err
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

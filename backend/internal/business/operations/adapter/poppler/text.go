// Package poppler implements PDF text extraction using the runtime's pdftotext.
package poppler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	operations "github.com/963sup/line-mini-app-bot/internal/business/operations/application"
)

const maxTextBytes = 8 << 20

// Extraction errors can be inspected with errors.Is without exposing PDF content.
var (
	ErrUnavailable = errors.New("pdftotext unavailable")
	ErrOutputLimit = errors.New("pdftotext output limit exceeded")
	ErrOpenPDF     = errors.New("pdftotext cannot open PDF")
	ErrOpenOutput  = errors.New("pdftotext cannot open output")
	ErrPermission  = errors.New("pdftotext PDF permission denied")
	ErrConversion  = errors.New("pdftotext conversion failed")
	ErrEncoding    = errors.New("pdftotext output is not UTF-8")
)

type Extractor struct{}

var _ operations.DocumentTextExtractor = (*Extractor)(nil)

// cappedOutput bounds memory even for PDFs whose text expands far beyond input.
type cappedOutput struct {
	bytes.Buffer
	exceeded bool
}

func (w *cappedOutput) Write(p []byte) (int, error) {
	if len(p) > maxTextBytes-w.Len() {
		w.exceeded = true
		return 0, io.ErrShortBuffer
	}
	return w.Buffer.Write(p)
}

func (*Extractor) ExtractText(ctx context.Context, doc operations.Document) (string, error) {
	if len(doc.Data) == 0 || len(doc.Data) > 8<<20 || !bytes.HasPrefix(doc.Data, []byte("%PDF-")) {
		return "", operations.ErrInvalidDocument
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// stdin/stdout avoid temporary files and untrusted filenames in arguments.
	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", "-enc", "UTF-8", "-eol", "unix", "-", "-")
	cmd.Stdin = bytes.NewReader(doc.Data)
	var output cappedOutput
	cmd.Stdout = &output
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", errors.Join(operations.ErrProvider, ctx.Err())
	}
	// Check the writer separately: process exit errors may mask a copy failure.
	if output.exceeded {
		return "", errors.Join(operations.ErrInvalidDocument, ErrOutputLimit)
	}
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", errors.Join(operations.ErrProvider, ErrUnavailable)
		}
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			switch exit.ExitCode() {
			case 1:
				return "", errors.Join(operations.ErrInvalidDocument, ErrOpenPDF)
			case 2:
				return "", errors.Join(operations.ErrProvider, ErrOpenOutput)
			case 3:
				return "", errors.Join(operations.ErrInvalidDocument, ErrPermission)
			}
		}
		return "", errors.Join(operations.ErrProvider, ErrConversion)
	}
	if !utf8.Valid(output.Bytes()) {
		return "", errors.Join(operations.ErrProvider, ErrEncoding)
	}
	return cleanText(output.String()), nil
}

// cleanText preserves layout and page boundaries; it does not normalize Unicode
// or collapse whitespace. C0/C1 controls other than LF, TAB and FF are removed.
func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\f' {
			return -1
		}
		return r
	}, text)
}

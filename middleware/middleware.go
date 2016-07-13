package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
)

const (
	CRLF            = "\r\n"
	COLONSPACE      = ": "
	SEMICOLON       = ";"
	CHECKSUMHEADERS = "X-Checksum-Headers"
	CHECKSUM        = "X-Checksum"
)

func ChecksumMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)

		// sortedKeys stores rec.Header() keys in lexicographic order.
		sortedKeys := make([]string, 0)
		for k, v := range rec.Header() {
			w.Header()[k] = v
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// canonResponse is the main buffer from which the checksum will be
		// generated.
		var canonResponse bytes.Buffer

		// canonResponse = "418\r\n"
		canonResponse.WriteString(strconv.Itoa(rec.Code))
		canonResponse.WriteString(CRLF)

		// checksumHeaders is the X-Checksum-Headers buffer which will be
		// appended to canonResponse before calculating the checksum.
		var checksumHeaders bytes.Buffer

		// checksumHeaders = "X-Checksum-Headers: "
		checksumHeaders.WriteString(CHECKSUMHEADERS)
		checksumHeaders.WriteString(COLONSPACE)

		// Iterate through rec.Header() in lexicographic order.
		for _, headerKey := range sortedKeys {
			// canonResponse + "Key: Value\r\n"
			canonResponse.WriteString(headerKey)
			canonResponse.WriteString(COLONSPACE)
			canonResponse.WriteString(rec.Header().Get(headerKey))
			canonResponse.WriteString(CRLF)

			// checksumHeaders + "Key;"
			checksumHeaders.WriteString(headerKey)
			checksumHeaders.WriteString(SEMICOLON)
		}

		// X-Checksum-Headers can't end with a semicolon, remove it.
		checksumHeaders.Truncate(checksumHeaders.Len() - 1)

		// checksumHeaders + "\r\n\r\n"
		checksumHeaders.WriteString(CRLF)
		checksumHeaders.WriteString(CRLF)

		// canonResponse + "X-Checksum-Headers: Key1;Key2;Key3;...\r\n"
		canonResponse.Write(checksumHeaders.Bytes())

		// canonResponse + content
		canonResponse.Write(rec.Body.Bytes())

		// Calculate checksum from canonResponse.
		hash := sha1.Sum(canonResponse.Bytes())
		checksum := hex.EncodeToString(hash[:])

		// Add X-Checksum header.
		w.Header().Set(CHECKSUM, checksum)

		w.WriteHeader(rec.Code)
		w.Write(rec.Body.Bytes())
	})
}

// Do not change this function.
func main() {
	var listenAddr = flag.String("http", ":8080", "addcanonResponses to listen on for HTTP")
	flag.Parse()

	http.Handle("/", ChecksumMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Foo", "bar")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Date", "Sun, 08 May 2016 14:04:53 GMT")
		msg := "Curiosity is insubordination in its purest form.\n"
		w.Header().Set("Content-Length", strconv.Itoa(len(msg)))
		fmt.Fprintf(w, msg)
	})))

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

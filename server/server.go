package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
)

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

//PrintProgress -
func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func imageHandler(w http.ResponseWriter, r *http.Request) {

	imageName := fmt.Sprintf("%s.img", r.RemoteAddr)

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("BootyImage")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	out, err := os.OpenFile(imageName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer out.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(file, counter)); err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("Beginning write of image [%s] to disk", imageName)

	w.WriteHeader(http.StatusOK)
}

// Serve will start the webserver for BOOTy
func main() {

	fs := http.FileServer(http.Dir("./images"))
	http.HandleFunc("/image", imageHandler)
	http.Handle("/images/", http.StripPrefix("/images/", fs))
	log.Println("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}

}
package image

// This package handles the pulling and management of images

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/thebsdbox/BOOTy/pkg/utils"
)

// Push - will take a local disk and copy an image to a remote server
func Push() {
	path := os.Getenv("CMDLINEPATH")
	if path == "" {
		path = utils.CmdlinePath
	}
	utils.ClearScreen()
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Starting BOOTy \n")
	fmt.Printf("\n\n")
	fmt.Printf("Parsing config from [%s]\n", path)
	src, dst, err := utils.ParseCmdLine(path)
	if err != nil {
		log.Fatalf("%v", err)
	}

	envSrc := os.Getenv("SRC")
	envDst := os.Getenv("DST")
	if envSrc == "" {
		//fmt.Printf("The \"SRC\" environment variable wasn't set")
	} else {
		src = envSrc
	}

	if envDst == "" {
		//fmt.Printf("The \"DST\" environment variable wasn't set")
	} else {
		dst = envDst
	}

	// TODO - consider timeouts

	client := &http.Client{}
	_, err = UploadMultipartFile(client, dst, "BootyImage", src)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("\n\n\n")

	fmt.Printf("Reading of disk [%s], and sending to [%s]", filepath.Base(src), dst)
	fmt.Printf("\n\n\n")

	fmt.Printf("\n\n\n")

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("\n\n\n\n")

	// TODO - reboot
	fmt.Println("This is where the push reboot happens :-D")
	realm.Reboot()

	time.Sleep(time.Second * 5)

}

//UploadMultipartFile -
func UploadMultipartFile(client *http.Client, uri, key, path string) (*http.Response, error) {
	body, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	mwriter := multipart.NewWriter(writer)
	req.Header.Add("Content-Type", mwriter.FormDataContentType())

	errchan := make(chan error)

	go func() {
		defer close(errchan)
		defer writer.Close()
		defer mwriter.Close()

		w, err := mwriter.CreateFormFile(key, path)
		if err != nil {
			errchan <- err
			return
		}

		in, err := os.Open(path)
		if err != nil {
			errchan <- err
			return
		}

		defer in.Close()

		if written, err := io.Copy(w, in); err != nil {
			errchan <- fmt.Errorf("error copying %s (%d bytes written): %v", path, written, err)
			return
		}

		if err := mwriter.Close(); err != nil {
			errchan <- err
			return
		}
	}()

	resp, err := client.Do(req)
	merr := <-errchan

	if err != nil || merr != nil {
		return resp, fmt.Errorf("http error: %v, multipart error: %v", err, merr)
	}

	return resp, nil
}

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

// Pull will pull an image and write it to local storage device
func Pull() {
	path := os.Getenv("CMDLINEPATH")
	if path == "" {
		path = utils.CmdlinePath
	}
	utils.ClearScreen()
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Starting BOOTy \n")
	fmt.Printf("\n\n")
	fmt.Printf("Parsing config from [%s]\n", path)
	src, dst, err := utils.ParseCmdLine(path)
	if err != nil {
		log.Fatalf("%v", err)
	}

	envSrc := os.Getenv("SRC")
	envDst := os.Getenv("DST")
	if envSrc == "" {
		//fmt.Printf("The \"SRC\" environment variable wasn't set")
	} else {
		src = envSrc
	}

	if envDst == "" {
		//fmt.Printf("The \"DST\" environment variable wasn't set")
	} else {
		dst = envDst
	}

	req, err := http.NewRequest("GET", src, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	var out io.Writer
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	out = f
	defer f.Close()

	fmt.Printf("\n\n\n")

	fmt.Printf("Beginning write of image [%s] to disk [%s]", filepath.Base(src), dst)
	fmt.Printf("\n\n\n")

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		log.Fatalf("%v", err)
	}

	count, err := io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("Error writing %d bytes to disk [%s] -> %v", count, dst, err)
	}
	fmt.Printf("\n\n\n")

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("\n\n\n\n")

	// TODO - reboot
	fmt.Println("This is where the reboot happens :-D")
	realm.Reboot()
	time.Sleep(time.Second * 5)
}
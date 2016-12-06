package main

import (
	"fmt"
	"io"
	"os"

	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"math"
	"strconv"
	"io/ioutil"
	"bytes"
	"strings"
	"bufio"
)

var path string = ""
var numberFiles int = 0;

func upload(c echo.Context) error {
	//-----------
	// Read file
	//-----------

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	path = file.Filename;
	splitFile(file.Filename); // split file in multiple files every 100KB

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully.</p>", file.Filename))
}

func search(c echo.Context) error {
	value := c.QueryParam("value")

	var line int = findString(path, value)

	//var found[] int;
	//for i := 0; i <= numberFiles; i++ {
		// go worker(i, jobs, results)
		//found[i] = findString("./part_" + i, value)
	//}

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>Text found line: %d</p>", line))
}

func splitFile(path string) {
	fileToBeChunked := path
	file, err := os.Open(fileToBeChunked)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close() // close the file at the end
	fileInfo, _ := file.Stat() // get file information

	var fileSize int64 = fileInfo.Size()
	const fileChunk = 100000 // 100 KB, change this to your requirement

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)
	numberFiles = int(totalPartsNum)

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize - int64(i * fileChunk))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)

		// write to disk
		fileName := "part_" + strconv.FormatUint(i, 10)
		_, err := os.Create(fileName)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// write/save buffer to disk
		ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)
		fmt.Println("Split to : ", fileName)
	}
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32 * 1024)
	count := 0
	lineSep := []byte{'\n'}
	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func findString(path string, pattern string) (int) {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)

	line := 1
	// https://golang.org/pkg/bufio/#Scanner.Scan
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), pattern) {
			//fmt.Println("Text found line: " + line)
			return line
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
	}

	return line
}

func worker(id int, jobs <-chan int, results chan <- int) {
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)

	e.GET("/search", search)

	e.Logger.Fatal(e.Start(":8080"))
}
// A simple tool to download daily bing wallpapers
// Takes two arguments
//
//	-s <image size in WxH format> (-s 1920x1080)
//	-o output file path
//
// and saves wallpaper of the day from Bing to output file
//
// $bingwp -s 1920x10180 -o /tmp/daily.jpg
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/image/draw"
)

const BING_API = "https://global.bing.com/HPImageArchive.aspx?format=js&idx=0&n=9&pid=hp&FORM=BEHPTB&uhd=1&uhdwidth=3840&uhdheight=2160&setmkt=en-US&setlang=en"
const BING_BASE_URL = "https://global.bing.com"

var szRegexp = regexp.MustCompile("(?P<Width>[0-9]+)[xX](?P<Height>[0-9]+)")

var ProxyError = errors.New("incorrect proxy address")
var HttpError = errors.New("incorrect http response")
var NoImagesError = errors.New("no images in bing response")
var NoImagePathError = errors.New("no image path in bing response")

var AppName string

func main() {
	AppName = filepath.Base(os.Args[0])
	var out, sz, proxy string
	flag.StringVar(&out, "o", "", "output file path")
	flag.StringVar(&sz, "s", "", "output image size")
	flag.StringVar(&proxy, "proxy", "", "optional proxy address")
	flag.Parse()
	if !validateSize(sz) || out == "" {
		fmt.Printf("incorrect params %s, %s\n", sz, out)
		printUsage()
		os.Exit(1)
	}
	out, err := filepath.Abs(out)
	if err != nil {
		fmt.Printf("incorrect output filepath %v\n", err)
		os.Exit(1)
	}
	client, err := createClient(proxy)
	if err != nil {
		fmt.Printf("can not create http client %v\n", err)
		os.Exit(1)
	}
	data, err := download(BING_API, client)
	if err != nil {
		fmt.Printf("can not dowload json data %v\n", err)
		os.Exit(1)
	}
	imgPath, err := findImagePath(data)
	if err != nil {
		fmt.Printf("can not find image path %v\n", err)
		os.Exit(1)
	}
	imgUrl := BING_BASE_URL + imgPath
	data, err = download(imgUrl, client)
	if err != nil {
		fmt.Printf("can not dowload image %v\n", err)
		os.Exit(1)
	}
	if err = saveAndBackupImage(data, out, sz); err != nil {
		fmt.Printf("can not save image %v\n", err)
		os.Exit(1)
	}
}

func validateSize(sz string) bool {
	return sz == "" || szRegexp.MatchString(sz)
}

func createClient(proxy string) (*http.Client, error) {
	if proxy == "" {
		return http.DefaultClient, nil
	}
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		// though proxy is optional but user spcified incorrect proxy address, return error
		return nil, err
	}
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}, nil
}

func download(uri string, client *http.Client) ([]byte, error) {
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w got status: %d", HttpError, resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func findImagePath(data []byte) (string, error) {
	type BingResponse struct {
		Images []struct {
			Url string `json:"url"`
		} `json:"images"`
	}
	js := BingResponse{}
	err := json.Unmarshal(data, &js)
	if err != nil {
		return "", err
	}
	if len(js.Images) == 0 {
		return "", NoImagesError
	}
	image := js.Images[0]
	if len(image.Url) == 0 {
		return "", NoImagePathError
	}
	return image.Url, nil
}

func saveAndBackupImage(data []byte, out, sz string) error {
	// save to tmp to not loose downloaded data
	fileDir := filepath.Dir(out)
	tmp, err := os.CreateTemp(fileDir, "tmp_*.jpg")
	if err != nil {
		return fmt.Errorf("can not create temporary image file %w", err)
	}
	defer tmp.Close()
	if sz != "" {
		if err = writeResizedImage(data, tmp, sz); err != nil {
			os.Remove(tmp.Name())
			return fmt.Errorf("can not save image %w", err)
		}
	} else {
		io.Copy(tmp, bytes.NewBuffer(data))
	}
	if fileExists(out) {
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		prevDateFileName := filepath.Join(fileDir, yesterday.Format("20060102.jpg"))
		if err := os.Rename(out, prevDateFileName); err != nil {
			return fmt.Errorf("can not backup file %s into %s. %w", out, prevDateFileName, err)
		}
	}
	if err = os.Rename(tmp.Name(), out); err != nil {
		return fmt.Errorf("can not rename temporary file %s into %s. %w", tmp.Name(), out, err)
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func writeResizedImage(data []byte, dst io.Writer, sz string) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	sizes := szRegexp.FindStringSubmatch(sz)
	w, _ := strconv.Atoi(sizes[1])
	h, _ := strconv.Atoi(sizes[2])
	minRatio := math.Min(
		float64(w)/float64(img.Bounds().Max.X),
		float64(h)/float64(img.Bounds().Max.Y))
	dstImage := image.NewRGBA(image.Rect(0, 0,
		int(float64(img.Bounds().Max.X)*minRatio),
		int(float64(img.Bounds().Max.Y)*minRatio)))
	draw.NearestNeighbor.Scale(dstImage, dstImage.Rect, img, img.Bounds(), draw.Over, nil)
	err = jpeg.Encode(dst, dstImage, nil)
	return err
}

func printUsage() {
	fmt.Printf("Usage: %s -s <image size> -o <output filename>\n"+
		" -o (required): Output filename (e.g., image.jpg or ~/pictures/image.jpg)\n\n"+
		" -s (optional): Image size in format WIDTHxHEIGHT (e.g., 1920x1080)\n"+
		" --proxy (optional) Proxy address if needed\n", AppName)
}

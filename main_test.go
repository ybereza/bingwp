package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestValidateSize(t *testing.T) {
	type testCase struct {
		in   string
		want bool
	}
	cases := []testCase{
		{"1024x768", true},
		{"1024X768", true},
		{"1x2", true},
		{"1024123X733368", true},
		{"1024", false},
		{"1024*768", false},
		{"1024,768", false},
		{"", true},
	}
	for _, tc := range cases {
		got := validateSize(tc.in)
		if tc.want != got {
			t.Fatalf("incorrect validation result in: %s, want %t, got %t", tc.in, tc.want, got)
		}
	}
}

func TestCreateClientWithProxy(t *testing.T) {
	type testCase struct {
		in   string
		want bool
	}
	cases := []testCase{
		{"http://localhost:1080", true},
		{"localhost:1080", true},
		{"http://my.proxy.com", true},
		{"http://my.proxy.com:8080", true},
		{"fw_someproxy:1080", false},
	}
	for _, tc := range cases {
		client, got := createClient(tc.in)
		if tc.want != (got == nil && client != nil) {
			t.Fatalf("incorrect result in: %s, want %v, got %v", tc.in, tc.want, (got == nil))
		}
	}
}

func TestCreateClientWithoutProxy(t *testing.T) {
	client, err := createClient("")
	if err != nil || client == nil {
		t.Fatalf("incorrect result. want no error bu got %v", err)
	}
	// In default client (without proxy) Transport is nil
	if client.Transport != nil {
		t.Fatalf("want client without custom transport but got %+v", client.Transport)
	}
}

func TestFindImagePathOK(t *testing.T) {
	testImagePath := "image.jpg"
	testData := fmt.Sprintf("{\"images\":[{\"url\":\"%s\", \"urlbase\":\"\"}]}", testImagePath)
	path, err := findImagePath([]byte(testData))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if path != testImagePath {
		t.Fatalf("want %s got %s", testImagePath, path)
	}
}

func TestFindImagePathNilData(t *testing.T) {
	_, err := findImagePath(nil)
	if err == nil {
		t.Fatalf("must return unmarshaling error but got nothing")
	}
}

func TestFindImagePathEmptyJSON(t *testing.T) {
	_, err := findImagePath([]byte("{}"))
	if !errors.Is(err, NoImagesError) {
		t.Fatalf("want %v but got %v", NoImagesError, err)
	}
}

func TestFindImagePathNoImagePath(t *testing.T) {
	bingResponse := "{\"images\":[{\"url\":\"\", \"urlbase\":\"\"}]}"
	_, err := findImagePath([]byte(bingResponse))
	if !errors.Is(err, NoImagePathError) {
		t.Fatalf("want %v but got %v", NoImagePathError, err)
	}
}

func TestFindImagePathEmptyImage(t *testing.T) {
	bingResponse := "{\"images\":[{}]}"
	_, err := findImagePath([]byte(bingResponse))
	if !errors.Is(err, NoImagePathError) {
		t.Fatalf("want %v but got %v", NoImagePathError, err)
	}
}

func TestSaveImageWithoutResize(t *testing.T) {
	testData := []byte{}
	testFileName := "test.jpg"
	err := saveAndBackupImage(testData, testFileName, "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !fileExists(testFileName) {
		t.Fatalf("output file %s was not created", testFileName)
	}
	os.Remove(testFileName)
}

func TestSaveImageWithoutResizeAlreadyExists(t *testing.T) {
	testData := []byte{}
	testFileName := "test.jpg"
	testFile, _ := os.Create(testFileName)
	testFile.Close()
	saveAndBackupImage(testData, testFileName, "")
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	prevDateFileName := yesterday.Format("20060102.jpg")
	if !fileExists(prevDateFileName) {
		t.Fatalf("prev date file %s was not backed up", prevDateFileName)
	}
	os.Remove(testFileName)
	os.Remove(prevDateFileName)
}

func TestSaveImageWithResize(t *testing.T) {
	testData := testJPG
	testFileName := "test.jpg"
	err := saveAndBackupImage(testData, testFileName, "1x1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !fileExists(testFileName) {
		t.Fatalf("output file %s was not created", testFileName)
	}
	os.Remove(testFileName)
}

func TestSaveIncorrectImageWithResize(t *testing.T) {
	testData := []byte{}
	testFileName := "test.jpg"
	err := saveAndBackupImage(testData, testFileName, "1x1")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

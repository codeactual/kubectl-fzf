package util

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestEncoding(t *testing.T) {
	data := "test"

	f, err := ioutil.TempFile("", "encoding")
	if err != nil {
		t.Fatalf("TempFile() error = %v", err)
	}
	defer os.Remove(f.Name())

	err = EncodeToFile(data, f.Name())
	if err != nil {
		t.Fatalf("EncodeToFile() error = %v", err)
	}

	var res string
	err = LoadGobFromFile(&res, f.Name())
	if err != nil {
		t.Fatalf("LoadGobFromFile() error = %v", err)
	}
	if res != "test" {
		t.Fatalf("expected \"test\", got %q", res)
	}
}

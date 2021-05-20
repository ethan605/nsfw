package crawler

import (
	"io/fs"
	"io/ioutil"
	"log"
)

type osFS struct {
	fs.FS
}

func (c osFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

// Instagram crawls data from instagram.com
func Instagram() {
	seeds := parseInstagramSeeds(osFS{})
	log.Println("Seeds:", seeds)
}

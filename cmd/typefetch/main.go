package main

import (
	"fmt"
	"github.com/csknk/polka-connect/core"
)

func main() {
	fmt.Println("vim-go")
}_

func GetLatestMetadata(t *testing.T) {
	c, err := polka-connect.NewConnection("http://localhost:9933")
	assert.NoError(t, err)

	meta, err := c.getLatestMetadata()
	assert.NoError(t, err)

	c.getEvents(meta)

}



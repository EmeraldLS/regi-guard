package utils

import (
	"fmt"
	"log"

	"github.com/mbndr/figlet4go"
)

func GeneratedAscii(text string) {

	// Create a new Figlet renderer
	renderer := figlet4go.NewAsciiRender()

	// Set the font
	err := renderer.LoadFont("standard.flf")
	if err != nil {
		log.Fatal("Error loading font:", err)
	}

	// Render the ASCII text
	renderedText, err := renderer.Render(text)
	if err != nil {
		log.Fatal("Error rendering text:", err)
	}

	// Print the rendered ASCII text
	fmt.Println(renderedText)
}

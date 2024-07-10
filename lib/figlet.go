package lib

import (
	"log"
	"os"
	"strings"

	"github.com/lukesampson/figlet/figletlib"
	"golang.org/x/exp/rand"
)

func Figlet(fontName string, text string) string {
	var (
		fontsDir = "fonts"
		// download figlet fonts from http://www.figlet.org/
		font *figletlib.Font
		err  error
	)

	if fontName == "random" {
		if files, err := os.ReadDir(fontsDir); err != nil {
			log.Printf("error reading directory %s: %v\n", fontsDir, err)
			return ""
		} else {
			fontName = files[rand.Intn(len(files))].Name()
		}
	}

	if font, err = figletlib.GetFontByName(fontsDir, fontName); err != nil {
		log.Printf("could not find font %s: %v\n", fontName, err)
		return ""
	}

	b := new(strings.Builder)
	figletlib.FPrintMsg(b, text, font, 80, font.Settings(), "left")
	return b.String()
}

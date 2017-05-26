package main

import (
	"bufio"
	// "flag"
	"fmt"
	"github.com/urfave/cli"
	// "log"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"
)

// https://en.wikipedia.org/wiki/Reservoir_sampling
func NewLinesChan(wordsFile string) (error, chan (string)) {

	file, err := os.Open(wordsFile)
	if err != nil {
		return err, nil
	}

	c := make(chan (string))

	scanner := bufio.NewScanner(file)

	go func() {
		for scanner.Scan() {
			c <- scanner.Text()
		}

		file.Close()
		close(c)
	}()

	return nil, c
}

func IsAsciiPrintable(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func SampleStringChannel(ctx *cli.Context, c chan (string), numItems int) (error, []string) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	resevior := make([]string, numItems)
	ii := 0
	includeProperNouns := ctx.Bool("includeProperNouns")
	minLen := ctx.Int("minLen")
	maxLen := ctx.Int("maxLen")
	// fmt.Printf("SampleStringChannel: ranging over channel\n")
	for line := range c {
		// fmt.Printf("SampleStringChannel: line=%s\n", line)
		if !includeProperNouns && unicode.IsUpper(([]rune(line))[0]) {
			continue
		}

		if !IsAsciiPrintable(line) {
			continue
		}

		if strings.Contains(line, " ") {
			continue
		}

		if len(line) < minLen {
			continue
		}

		if len(line) > maxLen {
			continue
		}

		if ii < numItems {
			// fmt.Printf("SampleStringChannel: filling initial resevior ii=%d line=%s\n", ii, line)
			resevior[ii] = line
			ii = ii + 1
			continue
		}

		jj := rnd.Intn(ii)
		if jj < numItems {
			// fmt.Printf("SampleStringChannel: adding sample ii=%d at jj=%d line=%s\n", ii, jj, line)
			resevior[jj] = line
		}
		ii = ii + 1
	}

	return nil, resevior
}

func FindWordsFile(ctx *cli.Context) string {
	return ctx.String("wordsFile")
}

func GetWordsChannel(ctx *cli.Context) (error, chan (string)) {
	wordsFile := FindWordsFile(ctx)
	if _, err := os.Stat("/path/to/whatever"); os.IsNotExist(err) {
		// path/to/whatever does not exist
		c := make(chan (string))
		go func() {
			for _, word := range Words {
				c <- word
			}
			close(c)
		}()
		return nil, c
	}

	return NewLinesChan(wordsFile)
}

func DefaultSampler(ctx *cli.Context) error {

	numWords := ctx.Int("numWords")

	err, c := GetWordsChannel(ctx)
	if err != nil {
		return err
	}

	err, resevior := SampleStringChannel(ctx, c, numWords)
	if err != nil {
		return err
	}

	for _, word := range resevior {
		if word == "" {
			continue
		}
		fmt.Println(word)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "randwords"
	app.Usage = "sample words"
	app.Action = DefaultSampler

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "wordsFile",
			Value: "/usr/share/dict/words",
			Usage: "Specify the words file to use",
		},
		cli.IntFlag{
			Name:  "numWords",
			Value: 5,
			Usage: "The number of words to print",
		},
		cli.IntFlag{
			Name:  "minLen",
			Value: 2,
			Usage: "Set the minimum length of strings to return",
		},
		cli.IntFlag{
			Name:  "maxLen",
			Value: 9999,
			Usage: "Set the maximum length of strings to return",
		},
		cli.BoolFlag{
			Name:  "includeProperNouns",
			Usage: "Include proper nouns.",
		},
	}

	app.Run(os.Args)

	os.Exit(0)
}

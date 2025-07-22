package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func GenerateRandomEDL(args []string) error {
	flagSet := flag.NewFlagSet("edl", flag.ContinueOnError)

	minimumShotLength := flagSet.Int("m", 7, "minimum length of any shot in seconds")
	maximumShotLength := flagSet.Int("n", 20, "maximum length of any shot in seconds")
	maximumCutsPerFile := flagSet.Int("f", 3, "maximum cuts taken for each file")
	outputFilePath := flagSet.String("o", "output.edl", "output file path")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	filenames := flagSet.Args()
	if len(filenames) == 0 {
		return errors.New("No files given.")
	}

	outputFile, err := os.OpenFile(*outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = fmt.Fprintf(
		outputFile,
		""+
			"TITLE: %s\n"+
			"FCM: NON-DROP FRAME\n\n",
		filepath.Base(*outputFilePath),
	)

	if err != nil {
		return err
	}

	outputOffset := 0
	clipNumber := 0
	for _, filename := range filenames {
		videoLength, err := getVideoLengthInSeconds(filename)
		if err != nil {
			panic(err)
		}

		cutStart := 0
		cutEnd := int(math.Min(
			float64(*minimumShotLength+1+rand.Intn(*maximumShotLength-*minimumShotLength)),
			float64(videoLength),
		))
		numCuts := 0
		maxCuts := 1 + rand.Intn(*maximumCutsPerFile)
		for cutStart < videoLength && numCuts < maxCuts {
			clipEdl := formatEdlEdit(clipNumber+1, cutStart, cutEnd, outputOffset, filename)
			_, err = fmt.Fprint(
				outputFile,
				clipEdl,
			)

			if err != nil {
				return err
			}

			outputOffset += cutEnd - cutStart
			cutStart = cutEnd + *maximumShotLength
			cutEnd = cutStart + 1 + rand.Intn(*maximumShotLength-*minimumShotLength)
			numCuts++
			clipNumber++
		}
	}

	return nil
}

func formatTimecode(seconds int) string {
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours()) % 100
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	return fmt.Sprintf("%.02d:%.02d:%.02d:01", h, m, s)
}

func formatEdlLine(clipNumber int, clipStart int, clipEnd int, offset int, isVideo bool) string {
	mediaType := "V"
	if !isVideo {
		mediaType = "A"
	}

	return fmt.Sprintf(
		"%.03d AX %s C %s %s %s %s",
		clipNumber,
		mediaType,
		formatTimecode(clipStart),
		formatTimecode(clipEnd),
		formatTimecode(offset),
		formatTimecode(offset+(clipEnd-clipStart)),
	)
}

func formatEdlEdit(clipNumber int, start int, end int, offset int, filename string) string {
	return fmt.Sprintf(
		""+
			"%s\n"+
			"%s\n"+
			"* FROM CLIP NAME: %s\n\n",
		formatEdlLine(clipNumber, start, end, offset, true),
		formatEdlLine(clipNumber, start, end, offset, false),
		filepath.Base(filename),
	)
}

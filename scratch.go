package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func scratch() {
	minimumShotTimeInSeconds := flag.Int("m", 7, "minimum length of any shot")
	maximumShotTimeInSeconds := flag.Int("x", 20, "maximum length of any shot")
	maximumCutsPerFile := flag.Int("f", 3, "maximum cuts taken for each file")
	outputFilePath := flag.String("o", "output.mp4", "output file path")

	flag.Parse()

	if _, err := os.Stat(*outputFilePath); err == nil {
		panic("Output file already exists")
	}

	outputBase := strings.TrimSuffix(*outputFilePath, filepath.Ext(*outputFilePath))

	edlOutputFilePath := fmt.Sprintf(
		"%s.edl",
		outputBase,
	)

	edlOutputFile, err := os.OpenFile(edlOutputFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer edlOutputFile.Close()

	if _, err = fmt.Fprintln(edlOutputFile, "# mpv EDL v0"); err != nil {
		panic(err)
	}

	cmxOutputFilePath := fmt.Sprintf(
		"%s.cmx.edl",
		outputBase,
	)

	cmxOutputFile, err := os.OpenFile(cmxOutputFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer cmxOutputFile.Close()

	if _, err = fmt.Fprintf(
		cmxOutputFile,
		""+
			"TITLE: %s\n"+
			"FCM: NON-DROP FRAME\n\n",
		outputBase,
	); err != nil {
		panic(err)
	}

	tempDir, err := os.MkdirTemp(".", "temp-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	filenames := flag.Args()
	concatFilePath := filepath.Join(tempDir, "concat.txt")
	concatFile, err := os.OpenFile(concatFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer concatFile.Close()

	cutNumber := 0
	totalOffsetInSeconds := 0
	for _, filename := range filenames {
		videoDurationInSeconds, err := getVideoLengthInSeconds(filename)
		if err != nil {
			panic(err)
		}

		cutStart := 0
		cutEnd := int(math.Min(
			float64(*minimumShotTimeInSeconds+1+rand.Intn(*maximumShotTimeInSeconds-*minimumShotTimeInSeconds)),
			float64(videoDurationInSeconds),
		))
		cutsForFile := 0
		maxCuts := 1 + rand.Intn(*maximumCutsPerFile)
		for cutStart < videoDurationInSeconds && cutsForFile < maxCuts {
			cutClip(
				filename,
				cutStart,
				cutEnd,
				totalOffsetInSeconds,
				cutNumber,
				tempDir,
				concatFile,
				edlOutputFile,
				cmxOutputFile,
			)

			totalOffsetInSeconds += cutEnd - cutStart
			cutStart = cutEnd + *maximumShotTimeInSeconds
			cutEnd = cutStart + 1 + rand.Intn(*maximumShotTimeInSeconds-*minimumShotTimeInSeconds)
			cutsForFile++
			cutNumber++
		}
	}

	concatCommand := exec.Command(
		"ffmpeg",
		"-f", "concat",
		"-i", concatFilePath,
		"-c", "copy",
		*outputFilePath,
	)

	fmt.Println(concatCommand)

	if output, err := concatCommand.Output(); err != nil {
		fmt.Println(string(output))
		panic(err)
	}

}

func cutClip(
	inputFilename string,
	start int,
	end int,
	clipOffset int,
	clipNumber int,
	tempDirPath string,
	concatFile io.Writer,
	edlOutputFile io.Writer,
	cmxOutputFile io.Writer,
) {
	cutOutputFile := fmt.Sprintf("%.04d-cut.mp4", clipNumber)
	cutOutputPath := filepath.Join(tempDirPath, cutOutputFile)

	cutCommand := exec.Command(
		"ffmpeg",
		"-ss", fmt.Sprintf("%d", start),
		"-to", fmt.Sprintf("%d", end),
		"-i", inputFilename,
		"-c:v", "copy",
		"-c:a", "aac",
		cutOutputPath,
	)

	fmt.Println(cutCommand)

	if err := cutCommand.Run(); err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintf(edlOutputFile, "%s,%d,%d\n", inputFilename, start, end); err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintf(
		cmxOutputFile,
		""+
			"%s\n"+
			"%s\n"+
			"* FROM CLIP NAME: %s\n\n",
		formatEdlLine(clipNumber+1, start, end, clipOffset, true),
		formatEdlLine(clipNumber+1, start, end, clipOffset, false),
		filepath.Base(inputFilename),
	); err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintf(concatFile, "file %s\n", cutOutputFile); err != nil {
		panic(err)
	}
}

package main

import (
	"flag"
	"fmt"
	"io"
	// "math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// func

func main() {
	// totalEditedTimeInSeconds := flag.Int("t", 300, "time of resulting edited video in seconds")
	minimumShotTimeInSeconds := flag.Int("m", 3, "minimum length of any shot")
	maximumShotTimeInSeconds := flag.Int("x", 20, "maximum length of any shot")
	maximumCutsPerFile := flag.Int("f", 3, "maximum cuts taken for each file")

	// useRandomOffset := flag.Bool("r", false, "use random offset for each clip")
	outputFilePath := flag.String("o", "output.mp4", "output file path")
	// tempDir := flag.String("t", "temp", "temp file directory")
	// outputFile := flag.String("o", "output.mp4", "output file path")

	flag.Parse()

	if _, err := os.Stat(*outputFilePath); err == nil {
		panic("Output file already exists")
	}

	commandsOutputFilePath := fmt.Sprintf(
		"%s_commands.txt",
		strings.TrimSuffix(*outputFilePath, filepath.Ext(*outputFilePath)),
	)

	commandsOutputFile, err := os.OpenFile(commandsOutputFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer commandsOutputFile.Close()

	// fmt.Println(*totalEditedTimeInSeconds)
	// fmt.Println(*minimumShotTimeInSeconds)
	// fmt.Println(*maximumShotTimeInSeconds)

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

	// videoTimes := make([]int, 0, len(filenames))
	// totalVideoTime := 0
	cutNumber := 0
	for _, filename := range filenames {
		decimalSecondsString, err := exec.Command(
			"ffprobe", "-v", "error",
			"-show_entries", "format=duration",
			"-of", "default=noprint_wrappers=1:nokey=1",
			filename,
		).Output()

		if err != nil {
			panic(err)
		}

		secondsFloat, err := strconv.ParseFloat(strings.TrimSpace(string(decimalSecondsString)), 64)
		if err != nil {
			panic(err)
		}
		videoDurationInSeconds := int(secondsFloat)

		cutStart := 0
		cutEnd := *minimumShotTimeInSeconds + 1 + rand.Intn(*maximumShotTimeInSeconds-*minimumShotTimeInSeconds)
		cutsForFile := 0
		maxCuts := 1 + rand.Intn(*maximumCutsPerFile)
		for cutStart < videoDurationInSeconds && cutsForFile < maxCuts {
			cutClip(
				filename,
				cutStart,
				cutEnd,
				cutNumber,
				tempDir,
				concatFile,
				commandsOutputFile,
			)

			cutStart = cutEnd + *maximumShotTimeInSeconds
			cutEnd = cutStart + 1 + rand.Intn(*maximumShotTimeInSeconds-*minimumShotTimeInSeconds)
			cutsForFile++
			cutNumber++
		}
		// if *useRandomOffset {
		// 	cutOffset = rand.Intn(int(math.Max(float64(videoDurationInSeconds-cutTime), 1)))
		// }

		// if cutTime >= videoDurationInSeconds {
		// 	cutTime = videoDurationInSeconds
		// 	cutOffset = 0
		// }
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

	// for i, filename := range filenames {

	// }

	// averageClipTime := int(*totalEditedTimeInSeconds / len(videoTimes))

	// if averageClipTime < *minimumShotTimeInSeconds {
	// fmt.Printf("Average clip time ")
	// }

	// fmt.Printf("Total video time: %d\n", totalVideoTime)
	// fmt.Printf("Average time per clip: %d\n", int(*totalEditedTimeInSeconds/len(videoTimes)))
	// for i := range filenames {
	// 	fmt.Printf("%s: %d seconds\n", filenames[i], videoTimes[i])
	// }
}

func cutClip(
	inputFilename string,
	start int,
	end int,
	clipNumber int,
	tempDirPath string,
	concatFile io.Writer,
	commandsOutputFile io.Writer,
) {
	cutOutputFile := fmt.Sprintf("%.04d-cut.mp4", clipNumber)
	cutOutputPath := filepath.Join(tempDirPath, cutOutputFile)

	// args := []string{}
	// // if start > 0 && end >= videoDurationInSeconds {
	// 	args = append(args,
	// 		"-ss", fmt.Sprintf("%d", start),
	// 		"-to", fmt.Sprintf("%d", end),
	// 	)
	// }
	// args = append(args,
	// 	"-i", filename,
	// 	"-c", "copy",
	// 	cutOutputPath,
	// )

	cutCommand := exec.Command(
		"ffmpeg",
		"-ss", fmt.Sprintf("%d", start),
		"-to", fmt.Sprintf("%d", end),
		"-i", inputFilename,
		"-c", "copy",
		cutOutputPath,
	)

	fmt.Println(cutCommand)
	fmt.Fprintln(commandsOutputFile, cutCommand.String())

	if err := cutCommand.Run(); err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintf(concatFile, "file %s\n", cutOutputFile); err != nil {
		panic(err)
	}
}

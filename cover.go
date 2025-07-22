package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

func GenerateRandomCoverImage(args []string) error {
	flagSet := flag.NewFlagSet("cover", flag.ContinueOnError)

	videoFilePath := flagSet.String("v", "", "video file to take thumbnail from (required)")
	mainText := flagSet.String("t", "", "main text")
	subText := flagSet.String("s", "", "sub text")
	figureImageFilePath := flagSet.String("f", "", "directory containing figure images")
	outputFilePath := flagSet.String("o", "output.png", "output file")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	videoLengthInSeconds, err := getVideoLengthInSeconds(*videoFilePath)
	if err != nil {
		return err
	}

	thumbnailOffsetSeconds := rand.Intn(videoLengthInSeconds)
	thumbnailOffsetMilliseconds := rand.Intn(1000) // this has a change of failing if the last second is chosen
	seekTimecode := formatSeekTimecode(thumbnailOffsetSeconds, thumbnailOffsetMilliseconds)

	fmt.Println(seekTimecode)

	err = exec.Command(
		"ffmpeg",
		"-i", *videoFilePath,
		"-ss", seekTimecode,
		"-frames:v", "1",
		*outputFilePath,
	).Run()

	if err != nil {
		return err
	}

	fmt.Println(videoFilePath, mainText, subText, figureImageFilePath, outputFilePath)

	return nil
}

func formatSeekTimecode(seconds, milliseconds int) string {
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours()) % 100
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	return fmt.Sprintf("%.02d:%.02d:%.02d.%.03d", h, m, s, milliseconds)
}

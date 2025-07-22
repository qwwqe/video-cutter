package main

import (
	"os/exec"
	"strconv"
	"strings"
)

func getVideoLengthInSeconds(filename string) (int, error) {
	secondsBytes, err := exec.Command(
		"ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filename,
	).Output()

	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(secondsBytes)), 64)
	if err != nil {
		return 0, err
	}

	return int(seconds), nil
}

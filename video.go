package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	var buf bytes.Buffer

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("couldn't run ffprobe: %w", err)
	}

	var probe struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(buf.Bytes(), &probe); err != nil {
		return "", fmt.Errorf("couldn't parse ffprobe output: %w", err)
	}

	if len(probe.Streams) == 0 {
		return "", fmt.Errorf("no streams found in video")
	}

	width := probe.Streams[0].Width
	height := probe.Streams[0].Height

	const tolerance = 0.1
	ratio := float64(width) / float64(height)

	switch {
	case math.Abs(ratio-16.0/9.0) < tolerance:
		return "16:9", nil
	case math.Abs(ratio-9.0/16.0) < tolerance:
		return "9:16", nil
	default:
		return "other", nil
	}
}

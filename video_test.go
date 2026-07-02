package main

import "testing"

func TestGetVideoAspectRatio(t *testing.T) {
	cases := map[string]string{
		"samples/boots-video-horizontal.mp4": "16:9",
		"samples/boots-video-vertical.mp4":   "9:16",
	}
	for path, want := range cases {
		got, err := getVideoAspectRatio(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if got != want {
			t.Errorf("%s: got %s, want %s", path, got, want)
		}
	}
}

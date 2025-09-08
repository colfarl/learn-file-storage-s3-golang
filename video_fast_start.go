package main

import "os/exec"


func processVideoForFastStart(filePath string) (string, error) {
	diskPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", diskPath)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return diskPath, nil
}

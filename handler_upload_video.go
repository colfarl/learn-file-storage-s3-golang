package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {	
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "video does not exist", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorized to access video", nil)
		return
	}

	const maxMemory = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("video")	
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to load file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")		
	
	fileType, _, err := mime.ParseMediaType(mediaType)
	if err != nil || fileType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "file must be an mp4 video", err)
		return
	}

	const tmpFileName = "tubely-upload.mp4"
	tmpFile, err := os.CreateTemp("", tmpFileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create file", err)
		return
	}	
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	_, err = io.Copy(tmpFile, file)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "could not copy data into file directory", err)
		return
	}
	
	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could reset tmp file data", err)
		return
	}

	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error making url", err)
		return
	}	

	if err = tmpFile.Close(); err != nil {
		respondWithError(w, 500, "could not close tmp file", err)
		return
	}

	ratio, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get aspect ratio of video", err)
		return
	}
	prefix := aspectRatioToPrefix(ratio)

		
	processedUploadFileName, err := processVideoForFastStart(tmpFile.Name()) 
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not preprocess video", err)
		return
	}

	uploadFile, err := os.Open(processedUploadFileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not reopen file", err)
		return
	}
	defer os.Remove(uploadFile.Name())
	defer uploadFile.Close()

	encoding := base64.RawURLEncoding.EncodeToString(bytes)
	fileKey := prefix + "/" + encoding + ".mp4"	

	objectParams := s3.PutObjectInput{
		Bucket : &cfg.s3Bucket,
		Key: &fileKey,
		Body: uploadFile,
		ContentType: &fileType,
	}

	_, err = cfg.client.PutObject(context.Background(), &objectParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to upload video to s3", err)
		return
	}
	
	publicURL := cfg.s3CfDistribution + fileKey
	video.VideoURL = &publicURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video", err)
		return
	}
	
	respondWithJSON(w, http.StatusOK, video)
}

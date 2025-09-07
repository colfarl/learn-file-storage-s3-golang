package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")	
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to load file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")	
	imageData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to read file data", err)
		return
	}
	
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "coould not retrieve video", err)
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorized to access video", nil)
	}
		
	encodedIMG := base64.StdEncoding.EncodeToString(imageData)
	thumbnailURL := fmt.Sprintf("data:%s;base64,%s", mediaType, encodedIMG)
	video.ThumbnailURL = &thumbnailURL 

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video", err)
	}

	respondWithJSON(w, http.StatusOK, video)
}

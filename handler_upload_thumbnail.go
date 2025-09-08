package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

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
	
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "coould not retrieve video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorized to access video", nil)
		return
	}
	
	fileType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not parse Content-Type", err)
		return
	}
	
	if fileType != "image/png" && fileType != "image/jpeg" {
		respondWithError(w, http.StatusBadRequest, "content must be an image png or jpeg", nil)
		return
	}

	fileExtension, err := mime.ExtensionsByType(fileType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not parse file type", err)
	}
	
	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error making url", err)
		return
	}
	
	encodedName := base64.RawURLEncoding.EncodeToString(bytes)
	fileName   := encodedName + fileExtension[0]
    diskPath   := filepath.Join(cfg.assetsRoot, fileName)
    publicPath := "/assets/" + fileName                   
	publicURL  := fmt.Sprintf("http://localhost:%s%s", cfg.port, publicPath)

	fileData, err := os.Create(diskPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create file", err)
		return
	}
	defer fileData.Close()	

	_, err = io.Copy(fileData, file)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "could not copy data into file directory", err)
		return
	}
		
	video.ThumbnailURL = &publicURL 

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video", err)
	}

	respondWithJSON(w, http.StatusOK, video)
}

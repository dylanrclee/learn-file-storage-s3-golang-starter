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
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing memory", err)
		return
	}

	file, fileheader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving file from key", err)
		return
	}
	defer file.Close()

	media_type := fileheader.Header.Get("Content-Type")
	image_bytes, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading thumbnail file bytes", err)
		return
	}
	image_base64 := base64.StdEncoding.EncodeToString(image_bytes)
	image_base64_url := fmt.Sprintf("data:%s;base64,%s", media_type, image_base64)

	video_data, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting video data", err)
		return
	}

	if video_data.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	video_data.ThumbnailURL = &image_base64_url
	err = cfg.db.UpdateVideo(video_data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating video data", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video_data)
}

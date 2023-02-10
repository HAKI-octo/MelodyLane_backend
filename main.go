package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/go-chi/chi"
)

const bucketName = "haki_melodylane" //REPLACE BUCKET NAME

var bkt *storage.BucketHandle

type Album struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Artist  string `json:"artist"`
	Content string `json:"content"`
}

const markdownTemplate = `
----
title: {{ .Title }}
artist: {{ .Artist }}
----

{{ .Content }}
`

func init() {

	client, err := storage.NewClient(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	//Making a bucket handle
	bkt = client.Bucket(bucketName)
}

func main() {

	r := chi.NewRouter()

	r.Post("/albums", createAlbumHandler)
	r.Delete("/albums/{id}", deleteAlbumHandler)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

}

func createAlbumHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	album := Album{
		ID:      r.PostFormValue("ID"),
		Title:   r.PostFormValue("Title"),
		Artist:  r.PostFormValue("Artist"),
		Content: r.PostFormValue("Content"),
	}

	// UPLOAD ALBUM METADATA
	objAlbumWriter := bkt.Object(album.ID + "/index.md").NewWriter(ctx)

	tmpl, err := template.New("album").Parse(markdownTemplate)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(objAlbumWriter, album)
	if err != nil {
		log.Fatal(err)
	}

	if err := objAlbumWriter.Close(); err != nil {
		log.Fatal(err)
	}

	objCoverWriter := bkt.Object(album.ID + "/cover.jpg").NewWriter(ctx)

	photo, _, _ := r.FormFile("cover")

	if _, err := io.Copy(objCoverWriter, photo); err != nil {
		log.Fatalf("Failed to write cover: %v", err)
	}

	if err := objCoverWriter.Close(); err != nil {
		log.Fatalf("Failed to close object: %v", err)
	}

	fmt.Fprintf(w, "Album created successfully")
}

func deleteAlbumHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	albumID := chi.URLParam(r, "id")

	obj := bkt.Object(albumID + "/index.md")
	if err := obj.Delete(ctx); err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Album deleted successfully")

}

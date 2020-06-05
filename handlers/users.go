package handlers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"fmt"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/oliviazhang/picture-retriever/libhttp"
	"github.com/oliviazhang/picture-retriever/models"
)

type Categories struct {
	Category []string `json:"dogs"`
}

type Record struct {
	Category string `json:"category"`
	URL      string `json:"url"`
}

// LoadPictures get random picture with category
func LoadPictures(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	jsonFile, err := os.Open(fmt.Sprintf("%s.json", category))
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}
	defer jsonFile.Close()

	var cates Categories

	db := r.Context().Value("db").(*sqlx.DB)
	imageFactory := models.NewImageFactory(db)

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &cates)
	for i, cate := range cates.Category {
		_, err := imageFactory.Save(nil, fmt.Sprintf("%s_%v", category, i), cate)
		if err != nil {
			libhttp.HandleErrorJson(w, err)
			return
		}
	}

	http.Redirect(w, r, "/", 302)
}

func SavePicture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := r.Context().Value("db").(*sqlx.DB)

	decoder := json.NewDecoder(r.Body)

	input := Record{}
	if err := decoder.Decode(&input); err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	_, err := models.NewImageFactory(db).Save(nil, input.Category, input.URL)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Perform login
	//NOTUSED_PostSave(w, r, input)
	http.Redirect(w, r, "/", 302)
}

// GetPicture get random picture with category
func GetPicture(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)

	category := r.URL.Query().Get("category")

	image, err := models.NewImageFactory(db).GetByCategoryLike(nil, category)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	imageBytes, _ := json.Marshal(image)
	_, _ = w.Write(imageBytes)
}

// PostLogin performs login.
func NOTUSED_PostSave(w http.ResponseWriter, r *http.Request, input Record) {
	w.Header().Set("Content-Type", "text/html")

	db := r.Context().Value("db").(*sqlx.DB)
	sessionStore := r.Context().Value("sessionStore").(sessions.Store)

	category := input.Category
	url := input.URL

	u := models.NewImageFactory(db)

	image, err := u.GetImageByCategoryAndURL(nil, category, url)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	session, _ := sessionStore.Get(r, "picture-retriever-session")
	session.Values["image"] = image

	err = session.Save(r, w)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func GetLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	sessionStore := r.Context().Value("sessionStore").(sessions.Store)

	session, _ := sessionStore.Get(r, "picture-retriever-session")

	delete(session.Values, "user")
	session.Save(r, w)

	http.Redirect(w, r, "/login", 302)
}

func PostPutDeleteImageID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	method := r.FormValue("_method")
	if method == "" || strings.ToLower(method) == "post" || strings.ToLower(method) == "put" {
		PutUsersID(w, r)
	} else if strings.ToLower(method) == "delete" {
		DeleteUsersID(w, r)
	}
}

func PutUsersID(w http.ResponseWriter, r *http.Request) {
	imageID, err := getIdFromPath(w, r)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	db := r.Context().Value("db").(*sqlx.DB)

	sessionStore := r.Context().Value("sessionStore").(sessions.Store)

	session, _ := sessionStore.Get(r, "picture-retriever-session")

	currentImage := session.Values["image"].(*models.ImageRow)

	if currentImage.ID != imageID {
		err := errors.New("Modifying other user is not allowed.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	category := r.FormValue("category")
	url := r.FormValue("url")

	u := models.NewImageFactory(db)

	currentImage, err = u.UpdateCategoryAndUrlById(nil, currentImage.ID, category, url)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Update currentUser stored in session.
	session.Values["image"] = currentImage
	err = session.Save(r, w)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func DeleteUsersID(w http.ResponseWriter, r *http.Request) {
	err := errors.New("DELETE method is not implemented.")
	libhttp.HandleErrorJson(w, err)
	return
}

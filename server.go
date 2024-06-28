package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var Genres = []string{"adventure", "drama", "detective", "comedy", "sci-fi"}

var comicGenresPageTemplate *template.Template
var comicInfoPageTemplate *template.Template
var mainPageTemplate *template.Template
var comicsFragmentTemplate *template.Template
var comicInfoLeftHalfTemplate *template.Template
var comicInfoRightHalfTemplate *template.Template

func indexOf(element string, arr []string) int {
	for i, v := range arr {
		if v == element {
			return i
		}
	}
	return -1
}

func encodeGenres(arr []string) int {
	result := 0
	for _, v := range arr {
		elInd := indexOf(v, Genres)
		if elInd == -1 {
			continue
		}
		result = result | (1 << elInd)
	}
	return result
}

func decodeGenres(n int) string {
	var result bytes.Buffer
	empty := true

	for i := 0; i < len(Genres); i++ {
		if (n>>i)&1 == 1 {
			if !empty {
				result.WriteString(", ")
			}

			result.WriteString(Genres[i])
			empty = false
		}
	}

	return result.String()
}

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func comicModelToComicView(comic comicModel) comicView {
	var result comicView
	result.Id = comic.Id
	result.Name = comic.Name
	result.Author = comic.Author
	result.Description = comic.Description
	result.Genres = decodeGenres(comic.Genres)

	return result
}

// func viewToComic(comicView comicView) comicModel {
// 	var result comicModel
// 	result.Id = comicView.Id
// 	result.Name = comicView.Name
// 	result.Author = comicView.Author
// 	result.Description = comicView.Description
// 	result.Genres = encodeGenres(strings.Split(comicView.Genres, ", "))

// 	return result
// }

func parseIdQueryParam(url *url.URL) (int, error) {
	idParam := url.Query().Get("id")

	comicId, err := strconv.Atoi(idParam)

	if err != nil {
		fmt.Println("error id field is not a number:", err)
	}

	return comicId, err
}

func executeTemplate(tmpl *template.Template, w http.ResponseWriter, viewData any) {
	if err := tmpl.Execute(w, viewData); err != nil {
		fmt.Println("error executing templates:", err)
		return
	}
}

func loadAllComicsHandler(w http.ResponseWriter, _ *http.Request) {
	var homePageViewData HomePageViewData

	comics := Map(loadAllComics(), comicModelToComicView)
	homePageViewData.Comics = comics
	homePageViewData.Genres = Genres
	executeTemplate(mainPageTemplate, w, homePageViewData)
}

func loadComicHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var comicModel comicModel
	var comicPageViewData ComicPageViewData
	var comicId int

	urlHeader := r.Header.Get("HX-Current-URL")

	url, err := url.Parse(urlHeader)
	if err != nil {
		fmt.Println("error parsing url:", err)
		goto ComicNotFound
	}
	comicId, err = parseIdQueryParam(url)
	if err != nil {
		fmt.Println("error parsing id:", err)
		goto ComicNotFound
	}

	comicModel, err = getComicById(comicId)
	if err != nil {
		goto ComicNotFound
	}

	comicPageViewData = ComicPageViewData{
		ChapterPublications: loadAllComicChapterPublications(comicId),
		Comic:               comicModelToComicView(comicModel),
	}
	executeTemplate(comicInfoPageTemplate, w, comicPageViewData)
	return

ComicNotFound:
	w.Write([]byte("Comic not found"))
}

func addComicHandler(w http.ResponseWriter, r *http.Request) {
	var comicModel comicModel
	//var homePageViewData HomePageViewData

	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error parsing form data:", err)
	}

	if r.PostForm.Has("name") {
		comicModel.Name = r.PostForm.Get("name")
	}
	if r.PostForm.Has("author") {
		comicModel.Author = r.PostForm.Get("author")
	}
	if r.PostForm.Has("decription") {
		comicModel.Description = r.PostForm.Get("description")
	}
	if r.PostForm.Has("genres") {
		comicModel.Genres = encodeGenres(r.PostForm["genres"])
	}

	saveComic(comicModel)

	comics := Map(loadAllComics(), comicModelToComicView)
	executeTemplate(comicsFragmentTemplate, w, comics)
	// homePageViewData.Comics = comics
	// homePageViewData.Genres = Genres
	// executeTemplate(mainPageTemplate, w, homePageViewData)
}

func deleteComicHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var comicId int

	urlHeader := r.Header.Get("HX-Current-URL")
	url, err := url.Parse(urlHeader)
	if err != nil {
		fmt.Println("Error parsing url:", err)
		goto Exit
	}
	comicId, err = parseIdQueryParam(url)
	if err != nil {
		fmt.Println("Error parsing id:", err)
		goto Exit
	}

	deleteComic(comicId)

Exit:
	w.Header().Set("HX-Push-Url", "/")
	loadAllComicsHandler(w, r)
}

func updateComicHandler(w http.ResponseWriter, r *http.Request) {
	var comicModel comicModel
	var comicView comicView
	var IdComic int

	var anon struct {
		hello string
	}

	anon.hello = "foo"
	fmt.Print(anon)

	urlHeader := r.Header.Get("HX-Current-URL")

	url, err := url.Parse(urlHeader)
	if err != nil {
		fmt.Println("Error parsing url:", err)
		goto RedirectToMainPage
	}

	IdComic, err = parseIdQueryParam(url)
	if err != nil {
		fmt.Println("Error finding id:", err)
		goto RedirectToMainPage
	}

	comicModel, err = getComicById(IdComic)
	if err != nil {
		fmt.Println("Error finding comic by id:", err)
		goto RedirectToMainPage
	}

	err = r.ParseForm()
	if err != nil {
		fmt.Println("Error parsing form data:", err)
		return
	}

	if r.PostForm.Has("name") {
		comicModel.Name = r.PostForm.Get("name")
	}
	if r.PostForm.Has("author") {
		comicModel.Author = r.PostForm.Get("author")
	}
	if r.PostForm.Has("decription") {
		comicModel.Description = r.PostForm.Get("description")
	}
	if r.PostForm.Has("genres") {
		comicModel.Genres = encodeGenres(r.PostForm["genres"])
	}

	saveComic(comicModel)
	comicView = comicModelToComicView(comicModel)

	executeTemplate(comicInfoLeftHalfTemplate, w, comicView)

	return

RedirectToMainPage:
	w.WriteHeader(303)
	w.Header().Set("Location", "/")
}

func filterComicsHandler(w http.ResponseWriter, r *http.Request) {
	filter := encodeGenres(r.URL.Query()["genres"])
	comics := Map(filterComics(filter), comicModelToComicView)
	executeTemplate(comicsFragmentTemplate, w, comics)
}

func loadComicGenresHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(comicGenresPageTemplate, w, Genres)
}

func addChapterPublicationHandler(w http.ResponseWriter, r *http.Request) {
	var chapterPublication chapterPublication

	//getting url in STRING!!
	urlHeader := r.Header.Get("HX-Current-URL")

	//parse url to url struct
	url, err := url.Parse(urlHeader)
	if err != nil {
		fmt.Println("Error parsing url:", err)
		goto Exit
	}

	chapterPublication.ComicId, err = parseIdQueryParam(url)
	if err != nil {
		fmt.Println("Error finding id:", err)
		goto Exit
	}

	// comicModel, err = getComicById(chapterPublication.ComicId)
	// if err != nil {
	// 	fmt.Println("Error finding comic by id:", err)
	// 	goto Exit
	// }

	err = r.ParseForm()
	if err != nil {
		fmt.Println("Error parsing form data:", err)
		goto Exit
	}

	if !r.PostForm.Has("no") {
		goto Exit
	}
	chapterPublication.ChapterNo, err = strconv.Atoi(r.PostForm.Get("no"))
	if err != nil {
		goto Exit
	}

	if r.PostForm.Has("name") {
		chapterPublication.Name = r.PostForm.Get("name")
	}

	saveChapterPublication(chapterPublication)

Exit:
	executeTemplate(comicInfoRightHalfTemplate, w, loadAllComicChapterPublications(chapterPublication.ComicId))
	// comicPageViewData.Comic = comicView
	// comicPageViewData.ChapterPublications = loadAllComicChapterPublications(comicPageViewData.Comic.Id)
	// executeTemplate(comicInfoPageTemplate, w, comicPageViewData)

}

func deleteChapterPublicationHandler() {

}

func main() {
	var err error
	var errTemp error

	comicGenresPageTemplate, errTemp = template.ParseFiles("templates/comic-genres.html")
	err = errors.Join(err, errTemp)

	comicInfoPageTemplate, errTemp = template.ParseFiles("templates/comic-info.html")
	err = errors.Join(err, errTemp)

	mainPageTemplate, errTemp = template.ParseFiles("templates/main-page.html")
	err = errors.Join(err, errTemp)

	comicsFragmentTemplate, errTemp = template.ParseFiles("templates/comics.html")
	err = errors.Join(err, errTemp)

	comicInfoLeftHalfTemplate, errTemp = template.ParseFiles("templates/comic-info-left-half.html")
	err = errors.Join(err, errTemp)

	comicInfoRightHalfTemplate, errTemp = template.ParseFiles("templates/comic-info-right-half.html")
	err = errors.Join(err, errTemp)

	if err != nil {
		log.Fatal("error parsing templates: ", err)
	}

	// Connect to the SQLite database
	db := openDatabase()
	defer db.Close()

	_, err = db.Exec(SQLMigration)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	} else {
		fmt.Println("DB initialized!")
	}

	http.HandleFunc("/main-page", loadAllComicsHandler)
	http.HandleFunc("/comic-info", loadComicHandler)
	http.HandleFunc("/add-comic", addComicHandler)
	http.HandleFunc("/delete-comic", deleteComicHandler)
	http.HandleFunc("/update-comic", updateComicHandler)
	http.HandleFunc("/filter-comics", filterComicsHandler)

	http.HandleFunc("/comic-genres", loadComicGenresHandler)

	http.HandleFunc("/add-chapter-publication", addChapterPublicationHandler)
	// http.HandleFunc("/delete-chapter-publication", deleteChapterPublicationHandler)

	fmt.Println("Starting echo server on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("error starting server:", err)
	}
}

package main

type comicModel struct {
	Id          int
	Name        string
	Author      string
	Description string
	Genres      int
}

type comicView struct {
	Id          int
	Name        string
	Author      string
	Description string
	Genres      string
}

type chapterPublication struct {
	Id        int
	ComicId   int
	Name      string
	ChapterNo int
}

type HomePageViewData struct {
	Genres []string
	Comics []comicView
}
type ComicPageViewData struct {
	ChapterPublications []chapterPublication
	Comic               comicView
}

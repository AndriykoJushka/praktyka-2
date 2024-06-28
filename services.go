package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func openDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal("error connecting to the database:", err)
	}

	return db
}

func saveComic(comic comicModel) {
	var err error
	db := openDatabase()
	defer db.Close()

	if comic.Id <= 0 {
		_, err = db.Exec(SQLInsertComicQuery, comic.Name, comic.Author, comic.Description, comic.Genres)
	} else {
		_, err = db.Exec(SQLUpdateComicQuery, comic.Name, comic.Author, comic.Description, comic.Genres, comic.Id)
	}

	if err != nil {
		fmt.Println("error saving comic:", err)
		return
	}
}

func saveChapterPublication(publication chapterPublication) {
	var err error
	db := openDatabase()
	defer db.Close()

	if publication.Id <= 0 {
		_, err = db.Exec(SQLInsertChapterPublicationQuery, publication.ComicId, publication.Name, publication.ChapterNo)
	} else {
		_, err = db.Exec(SQLUpdateChapterPublicationQuery, publication.ComicId, publication.Name, publication.ChapterNo, publication.Id)
	}

	if err != nil {
		fmt.Println("error saving chapter publication:", err)
		return
	}
}

func deleteComic(id int) error {
	db := openDatabase()
	defer db.Close()

	_, err := db.Exec(SQLDeleteComicQuery, id)
	if err != nil {
		fmt.Println("error deleting comic:", err)
	}

	return err
}

func deleteChapterPublication(id int) error {
	db := openDatabase()
	defer db.Close()

	_, err := db.Exec(SQLDeleteChapterPublicationQuery, id)
	if err != nil {
		fmt.Println("Error deleting data:", err)
	}

	return err
}

func loadAllComicChapterPublications(comicId int) []chapterPublication {
	db := openDatabase()
	defer db.Close()

	rows, err := db.Query(SQLLoadAllComicChapterPublicationsQuery, comicId)
	if err != nil {
		fmt.Println("error loading all comic chapter publications:", err)
		return nil
	}
	defer rows.Close()

	var publications []chapterPublication

	for rows.Next() {
		var publication chapterPublication
		err := rows.Scan(&publication.Id, &publication.ComicId, &publication.Name, &publication.ChapterNo)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		publications = append(publications, publication)
	}

	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return publications
}

func loadAllComics() []comicModel {
	db := openDatabase()
	defer db.Close()

	rows, err := db.Query(SQLLoadAllComicsQuery)
	if err != nil {
		fmt.Println("error loading all comics:", err)
		return nil
	}
	defer rows.Close()

	var comics []comicModel

	for rows.Next() {
		var comic comicModel
		err := rows.Scan(&comic.Id, &comic.Name, &comic.Author, &comic.Description, &comic.Genres)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		comics = append(comics, comic)
	}

	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return comics
}

func getComicById(id int) (comicModel, error) {
	var err error
	var rows *sql.Rows
	var comic comicModel

	db := openDatabase()
	defer db.Close()

	rows, err = db.Query(SQLGetComicByIdQuery, id)
	if err != nil {
		goto Exit
	}

	defer rows.Close()

	if !(rows.Next()) {
		err = errors.New("comic not found")
		goto Exit
	}

	err = rows.Scan(&comic.Id, &comic.Name, &comic.Author, &comic.Description, &comic.Genres)
	err = errors.Join(err, rows.Err())

Exit:
	return comic, err
}

func filterComics(filter int) []comicModel {
	db := openDatabase()
	defer db.Close()

	rows, err := db.Query(SQLLoadAllComicsFilteredByGenresQuery, filter, filter)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	defer rows.Close()

	var comics []comicModel

	for rows.Next() {
		var comic comicModel
		err := rows.Scan(&comic.Id, &comic.Name, &comic.Author, &comic.Description, &comic.Genres)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		comics = append(comics, comic)
	}

	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return comics
}

package main

const SQLMigration string = `
	create table if not exists comic(id integer primary key autoincrement, name text not null unique, author text, description text, genres integer);
	create table if not exists publication(id integer primary key, id_comic integer not null, name text, chapter_no integer not null, foreign key(id_comic) references comic(id));
`

// Comic table operations
const SQLInsertComicQuery = `
	insert into comic(name,author,description,genres) values(?,?,?,?);
`
const SQLUpdateComicQuery = `
	update comic set name = ?, author = ?, description = ?, genres = ? where id = ?;
`
const SQLDeleteComicQuery = `
	delete from comic where id = ?;
`
const SQLLoadAllComicsQuery = `
	select * from comic; 
`
const SQLGetComicByIdQuery = `
	select * from comic where id = ?;
`
const SQLLoadAllComicsFilteredByGenresQuery = `
	select * from comic where ((? & genres) = ?);
`

// Publication operations
const SQLInsertChapterPublicationQuery = `
	insert into publication(id_comic,name,chapter_no) values(?,?,?);
`

const SQLUpdateChapterPublicationQuery = `
	update publication set id_comic = ?, name = ?, chapter_no = ? where id = ?;
`

const SQLDeleteChapterPublicationQuery = `
	delete from publication where id = ?;
`

const SQLLoadAllComicChapterPublicationsQuery = `
	select * from publication where id_comic = ? order by chapter_no asc; 
`

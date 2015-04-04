package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"vallon.me/shortening"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	_ "github.com/mattn/go-sqlite3"
)

var (
	APIKey    = ""
	DropBoxID = ""
)

func add(c web.C, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	key := r.PostForm.Get("key")
	if key != APIKey {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	file := r.PostForm.Get("file")
	if file == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	res, err := insert.Exec("https://dl.dropboxusercontent.com/u/" + DropBoxID + "/" + file)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", shortening.Encode(uint64(id)))
}

func view(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := shortening.Decode([]byte(c.URLParams["id"]))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	row := find.QueryRow(id)

	var url string
	if err := row.Scan(&url); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

var insert, find *sql.Stmt

func main() {
	goji.Serve()
}

func init() {
	goji.Post("/s", add)
	goji.Get("/s/:id", view)

	db, err := sql.Open("sqlite3", "./dropfile.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS links(id INTEGER PRIMARY KEY AUTOINCREMENT, url TEXT)`)
	if err != nil {
		panic(err)
	}

	insert, err = db.Prepare(`INSERT INTO links(url) VALUES(?)`)
	if err != nil {
		panic(err)
	}

	find, err = db.Prepare(`SELECT url FROM links WHERE id = ?`)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

type Post struct {
	Id    int
	Name  string
	Text  string
	Prise string
}

var database *sql.DB

// Создание постов

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		name := r.FormValue("name")
		text := r.FormValue("text")
		prise := r.FormValue("prise")

		if name == "" || text == "" || prise == "" {
			fmt.Fprintf(w, "Не все поля заполнены")
		} else {

			_, err = database.Exec("insert into `post` (name, text, prise) values (?, ?,?)",
				name, text, prise)
		}
		if err != nil {
			log.Println(err)
		}
		http.Redirect(w, r, "/", 301)
	} else {
		http.ServeFile(w, r, "template/create.html")
	}
}

// Переносим на страницу изменения поста

func EditPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	row := database.QueryRow("select * from `post` where id = ?", id)
	post := Post{}
	err := row.Scan(&post.Id, &post.Name, &post.Text, &post.Prise)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		tmpl, _ := template.ParseFiles("template/edit.html")
		tmpl.Execute(w, post)
	}
}

// Сохранение данных в бд после изменения

func EditHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	id := r.FormValue("id")
	name := r.FormValue("name")
	text := r.FormValue("text")
	prise := r.FormValue("prise")
	_, err = database.Exec("update `post` set name=?, text=?,prise=? where id = ?",
		name, text, prise, id)

	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", 301)
}

// Удаление данных

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := database.Exec("delete from `post` where id = ?", id)
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/", 301)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := database.Query("select * from `post`")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	posts := []Post{}

	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Id, &p.Name, &p.Text, &p.Prise)
		if err != nil {
			fmt.Println(err)
			continue
		}
		posts = append(posts, p)
	}

	tmpl, _ := template.ParseFiles("template/index.html")
	tmpl.Execute(w, posts)
}

func main() {

	db, err := sql.Open("mysql", "root:root@/goland")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	database = db
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/create/", CreateHandler)
	router.HandleFunc("/edit/{id:[0-9]+}", EditPage).Methods("GET")
	router.HandleFunc("/edit/{id:[0-9]+}", EditHandler).Methods("POST")
	router.HandleFunc("/delete/{id:[0-9]+}", DeleteHandler)

	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}

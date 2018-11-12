package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(instanceConnName string, dbname string, user string, password string) {
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		instanceConnName,
		dbname,
		user,
		password)

	db, err := sql.Open("cloudsqlpostgres", dsn)
	a.DB = db
	if err != nil {
		log.Fatal(err)
	}

	err = a.DB.Ping()
	if err != nil {
		panic(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()

	fmt.Println("Successfully connected to DB!")
}

func (a *App) Run(addr string) {
	headersOk := handlers.AllowedHeaders([]string{"Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})
	fmt.Printf("Sever running on port 8080")
	log.Fatal(http.ListenAndServe(addr, handlers.CORS(originsOk, headersOk, methodsOk)(a.Router)))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/healthcheck", healthcheck).Methods("GET")
	a.Router.HandleFunc("/people", a.getPeople).Methods("GET")
	a.Router.HandleFunc("/people/{id}", a.getPerson).Methods("GET")
	a.Router.HandleFunc("/people", a.createPerson).Methods("POST")
}

func (a *App) getPeople(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query("SELECT uid, firstname, lastname, email FROM users")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	people := []Person{}

	for rows.Next() {
		var (
			id        string
			firstname string
			lastname  string
			email     string
		)
		err := rows.Scan(&id, &firstname, &lastname, &email)
		if err != nil {
			log.Fatal(err)
		}
		var p Person
		p.ID = id
		p.Firstname = firstname
		p.Lastname = lastname
		p.Email = email

		people = append(people, p)
		log.Printf("%v: %s %s %s", id, firstname, lastname, email)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	response, _ := json.Marshal(people)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *App) getPerson(w http.ResponseWriter, r *http.Request) {
	variables := mux.Vars(r)
	id := variables["id"]
	p := Person{}
	p.ID = id

	row := a.DB.QueryRow("SELECT uid, firstname, lastname, email FROM users WHERE uid=$1", p.ID)
	err := row.Scan(&p.ID, &p.Firstname, &p.Lastname, &p.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			response, _ := json.Marshal(map[string]string{"error": "Product not found"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(response)
			return
		}
		log.Fatal(err)
	}

	response, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *App) createPerson(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	p := Person{}
	err := decoder.Decode(&p)
	if err != nil {
		response, _ := json.Marshal(map[string]string{"error": "Invalid request payload"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}
	log.Println("Person:", p)
	defer r.Body.Close()

	row := a.DB.QueryRow(
		"INSERT INTO users(firstname, lastname, email) VALUES ($1, $2, $3) RETURNING uid",
		p.Firstname, p.Lastname, p.Email)
	err = row.Scan(&p.ID)

	if err != nil {
		response, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(response)
		return
	}
	response, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("LOOKING GOOD!")
}

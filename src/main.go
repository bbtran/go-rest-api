package main

type Person struct {
	ID        string `json:"id,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Email     string `json:"email,omitempty"`
}

var people []Person

func main() {
	a := App{}
	a.Initialize("my-project-1471704628324:us-west1:pg-test-db", "test-pg-db", "postgres", "Benaivainc!")

	a.Run(":8080")
}

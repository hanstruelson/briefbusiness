package main

import (
	"appeals/db"
	"appeals/routes/user"
	"fmt"
	"net/http"
)

func main() {
	err := db.InitDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		panic(err)
	}
	err = db.RunMigrations()
	if err != nil {
		fmt.Println("Error running migrations:", err)
		panic(err)
	}
	http.HandleFunc("/api/user/create", user.CreateUser)
	http.ListenAndServe(":8085", nil)
}

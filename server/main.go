package main

import (
	"appeals/db"
	"appeals/routes/draft"
	"appeals/routes/user"
	"fmt"
	"net/http"
	"strings"
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
	
	http.HandleFunc("/api/drafts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			draft.ListDrafts(w, r)
		} else if r.Method == http.MethodPost {
			draft.SaveDraft(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/drafts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/drafts/backups/") {
			draft.GetDraftBackup(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/pdf") {
			draft.GeneratePDFHandler(w, r)
			return
		}
		draft.GetDraft(w, r)
	})

	http.ListenAndServe(":8085", nil)
}

package user

import (
	"appeals/db"
	"appeals/utilities"
	"fmt"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	success, data := utilities.DecodeJSON(w, r, &CreateUserData{})
	if !success {
		return
	}
	result, err := db.DB.Exec("INSERT INTO User (Email) VALUES (?)", data.Email)
	fmt.Println(result)
	if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Error creating user")
		return
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Error retrieving last insert ID")
		return
	}
	createUserResponse := CreateUserResponse{
		UserID:      int(lastInsertId),
		Email:       data.Email,
		AccessToken: utilities.GenerateAccessToken(int(lastInsertId)),
	}
	utilities.SendJSON(w, &createUserResponse)
	println(data.Email)
}

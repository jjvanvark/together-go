package routing

import (
	"encoding/json"
	"log"
	"maus/together-go/database"
	"net/http"
)

func handleStart(rw http.ResponseWriter, req *http.Request, user *database.User) {
	collaborators, err := db.GetCollaborators(user.Collaborators)
	if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	collaborating, err := db.GetCollaborating(user.Id)
	if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(rw).Encode(struct {
		User          *database.User  `json:"user"`
		Collaborators []database.User `json:"collaborators"`
		Collaborating []database.User `json:"collaborating"`
	}{
		user,
		collaborators,
		collaborating,
	})

	if err != nil {
		log.Println(err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
}

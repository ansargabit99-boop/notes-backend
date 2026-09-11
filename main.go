package main
//
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
)
type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}
var pool *pgxpool.Pool
func main() {
	var err error
	pool,err = pgxpool.New(context.Background(),os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("unable to connect database",err)
	}
	defer pool.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /notes",getNotes)
	mux.HandleFunc("POST /notes",addNotes)
	mux.HandleFunc("DELETE /notes/:id",deleteNotes)
	mux.HandleFunc("PATCH /notes/:id",changeNote)
	log.Println("server running on localhost:3000")
	handler := cors.New(cors.Options{
    AllowedOrigins: []string{"http://localhost:5173"},
    AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE"},
    AllowedHeaders: []string{"Content-Type"},
}).Handler(mux)

http.ListenAndServe(":3000", handler)
}
func getNotes(w http.ResponseWriter,r *http.Request) {
	rows,err := pool.Query(r.Context(),"SELECT * FROM notes") // whar type of function is thsi 
	if err!=nil {
		http.Error(w,"something went wrong",http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var notes []Note /// i still dont understand why is this here 
	for rows.Next() {//what does rows.Next do here is it like i++?
		var n Note
		err := rows.Scan(&n.ID,&n.Title,&n.Body)////what does scan mean here 
		if err != nil {
			http.Error(w,"failed to fetch",http.StatusInternalServerError)
			return
		}
		notes = append(notes,n)
	}
	w.Header().Set("Content-Type","application/json")//this line
	json.NewEncoder(w).Encode(notes)///this line

}
func addNotes(w http.ResponseWriter, r *http.Request) {
	var n Note//this line why we declare that here
	err:=json.NewDecoder(r.Body).Decode(&n)//this line and & why we eveb n use it why are we even giveing this varable of err not simple
	if err != nil  {
		http.Error(w,"something went wrong",http.StatusBadRequest)
		return
	}
	err=pool.QueryRow(r.Context(),"INSERT INTO notes (title,body) VALUES($1,$2) RETURNING id",n.Title,n.Body).Scan(&n.ID)//what does scan mean here and why query row not just query
	if err != nil {
		http.Error(w,"something went wrong",http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/json")//what does header mean here and set do
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)/// here am i sending just id here or whole note 
	
}

func deleteNotes(w http.ResponseWriter,r *http.Request) {
	var n Note
	id:= r.PathValue("id")
	idInt,err := strconv.Atoi(id)
	if err != nil {
		http.Error(w,"InvalidId",http.StatusBadRequest)
		return
	}
	err = pool.QueryRow(r.Context(),"DELETE FROM notes WHERE id=$1 RETURNING id",idInt).Scan(&n.ID)
	if err != nil {
		http.Error(w,"something went wrong try again",http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
type UpdateInput struct {
	Title string `json:"title"`
	Body string `jsoon:"title"`
}
func changeNote(w http.ResponseWriter,r *http.Request) {
	var n Note
	id := r.PathValue("id")
	intId,err := strconv.Atoi(id)
	if err != nil {
		http.Error(w,"invalidId",http.StatusBadRequest)
		return
	}
	var input UpdateInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w,"invalid header",http.StatusBadRequest)
		return
	}
	err = pool.QueryRow(r.Context(), `
	UPDATE notes SET 
		title = COALESCE($1,title),
		body = COALESCE($2,body) WHERE id=$3 RETURNING *
	`,input.Title,input.Body,intId).Scan(&n.ID,&n.Title,&n.Body)
	if err != nil {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n)
}
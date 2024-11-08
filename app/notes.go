package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"

	_ "embed"
)

//go:embed notes.html
var notesTplSrc string

//go:embed notes.css
var notesCss string

var notesTpl = template.Must(template.New("notes").Parse(notesTplSrc))

type note struct {
	Id             int
	Title, Content string
}

type notesHandler struct {
	db *sql.DB
}

func scanNote(rows *sql.Rows) (nt note, err error) {
	err = rows.Scan(&(nt.Id), &(nt.Title), &(nt.Content))
	return nt, err
}

func (nh *notesHandler) initialize(ctx context.Context) error {
	must(nh.db.Exec(`CREATE TABLE IF NOT EXISTS notes(id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, content TEXT)`))
	nts, err := nh.getNotes(ctx)
	if err != nil {
		return err
	}
	if len(nts) == 0 {
		log.Println("No notes found, initializing...")
		if err := nh.putNote(ctx, note{Title: "Salutations", Content: "Hello, World!"}); err != nil {
			return err
		}
	}
	return nil
}

func (nh *notesHandler) getNotes(ctx context.Context) ([]note, error) {
	// Retrieve notes
	rows, err := nh.db.QueryContext(ctx, `SELECT * FROM notes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []note
	for rows.Next() {
		nt, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, nt)
	}
	if err := rows.Err(); err != nil {
		return nil, rows.Err()
	}
	return notes, nil
}

func (nh *notesHandler) putNote(ctx context.Context, nt note) error {
	_, err := nh.db.ExecContext(ctx, `INSERT INTO notes(title, content) VALUES('`+nt.Title+`', '`+nt.Content+`')`)
	return err
}

func (nh *notesHandler) getNote(ctx context.Context, id int) (note, error) {
	rows, err := nh.db.QueryContext(ctx, `SELECT * FROM notes WHERE id = `+strconv.Itoa(id))
	if err != nil {
		return note{}, err
	}
	if !rows.Next() {
		return note{}, errors.New("not found")
	}
	nt, err := scanNote(rows)
	if err != nil {
		return nt, err
	}
	return nt, rows.Err()
}

func (nh *notesHandler) deleteNote(ctx context.Context, id int) error {
	_, err := nh.db.ExecContext(ctx, `DELETE FROM notes WHERE id = `+strconv.Itoa(id))
	return err
}

func Notes() http.Handler {
	db := must(sql.Open("sqlite", "./notes.db"))
	nh := &notesHandler{db}
	if err := nh.initialize(context.Background()); err != nil {
		log.Fatalf("Cannot initialize notes: %v", err)
	}

	n := http.NewServeMux()

	// Home for the note page
	n.HandleFunc("/notes/", func(w http.ResponseWriter, r *http.Request) {
		notes, err := nh.getNotes(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		// Write the template with the notes
		if err = notesTpl.Execute(w, notes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
		}
	})
	n.HandleFunc("/notes/notes.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/css")
		io.WriteString(w, notesCss)
	})

	// Add notes
	n.HandleFunc("/notes/add", func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		content := r.FormValue("content")
		if title == "" || content == "" {
			io.WriteString(w, `<html>
	The title and content cannot be empty. <a href="/notes">Go back</a>
	</html>`)
			return
		}
		if err := nh.putNote(r.Context(), note{
			Title:   title,
			Content: content,
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		http.Redirect(w, r, "/notes", http.StatusTemporaryRedirect)
	})
	// Delete notes
	n.HandleFunc("/notes/delete", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Fprintf(w, `<html>
	Invalid note ID: %v <a href="/notes">Go back</a>
	</html>`, r.FormValue("id"))
			return
		}
		if err := nh.deleteNote(r.Context(), id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}
		http.Redirect(w, r, "/notes", http.StatusTemporaryRedirect)
	})

	return n
}
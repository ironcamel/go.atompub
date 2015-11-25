package main

import (
	"database/sql"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ironcamel/go.atom"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/unrolled/render" // or "gopkg.in/unrolled/render.v1"
)

var r = render.New()
var db *sql.DB

func main() {
	var err error

	var port int
	envPort := os.Getenv("GO_ATOMPUB_PORT")
	if envPort == "" {
		port = 8000
	} else {
		if port, err = strconv.Atoi(envPort); err != nil {
			log.Fatal("Invalid port value: ", envPort)
		}
	}
	flag.IntVar(&port, "port", port, "the port")

	var dsn string
	flag.StringVar(&dsn, "dsn", "", "the database dsn")
	flag.Parse()

	if dsn == "" {
		dsn = os.Getenv("GO_ATOMPUB_DSN")
		if dsn == "" {
			log.Fatal("--dsn flag or GO_ATOMPUB_DSN env var is required")
		}
	}

	if db, err = sql.Open("postgres", dsn); err != nil {
		log.Fatal("Could not open db: ", err)
	}

	log.Println(time.Now().Format(time.RFC3339))

	router := mux.NewRouter()
	router.HandleFunc("/feeds/{feed}", getFeed).Methods("GET")
	router.HandleFunc("/feeds/{feed}", addEntry).Methods("POST")
	router.HandleFunc("/feeds/{feed}/entries/{entry}", getEntry).Methods("GET")
	log.Println("Listening on port", port)
	http.ListenAndServe(fmt.Sprint(":", port), router)
}

func getFeed(w http.ResponseWriter, req *http.Request) {
	feedTitle := mux.Vars(req)["feed"]
	feedPtr, err := findFeed(feedTitle)
	if err != nil {
		if err == sql.ErrNoRows {
			r.Text(w, 404, "No such feed")
			return
		} else {
			r.Text(w, 500, fmt.Sprint("Failed to get feed: ", err))
			return
		}
	}
	if err := populateFeed(feedPtr, req.FormValue("start_after")); err != nil {
		r.Text(w, 500, fmt.Sprint("Failed to construct feed: ", err))
		return
	}
	resXML(w, feedPtr)
}

func getEntry(w http.ResponseWriter, req *http.Request) {
	entryPtr, err := findEntry(mux.Vars(req)["entry"])
	if err != nil {
		if err == sql.ErrNoRows {
			r.Text(w, 404, "No such entry")
			return
		} else {
			r.Text(w, 500, fmt.Sprint("Failed to get entry: ", err))
			return
		}
	}
	resXML(w, entryPtr)
}

func findEntry(id string) (*atom.XMLEntry, error) {
	row := db.QueryRow(
		`select id, title, content, order_id from atom_entry where id = $1`, id)
	return entryFromRow(row)
}

func populateFeed(feed *atom.XMLFeed, startAfter string) error {
	minId := 0
	if startAfter != "" {
		entry, err := findEntry(startAfter)
		if err == nil {
			minId = *entry.IntId
		}
	}
	rows, err := db.Query(
		`select id, title, content, order_id
		from atom_entry
		where feed_title = $1
		and order_id > $2
		order by order_id
		limit 100`,
		feed.Title.Raw, minId,
	)
	defer rows.Close()
	if err != nil {
		return err
	}
	var entries []atom.XMLEntry
	for rows.Next() {
		entry, err := entryFromRow(rows)
		if err != nil {
			return err
		}
		entries = append(entries, *entry)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	feed.Entries = entries
	return nil
}

func resXML(w http.ResponseWriter, data interface{}) {
	var type1 string
	namespace := "http://www.w3.org/2005/Atom"
	switch x := data.(type) {
	case *atom.XMLEntry:
		type1 = "entry"
		x.Namespace = &namespace
	case *atom.XMLFeed:
		type1 = "feed"
		x.Namespace = &namespace
	}
	cont := fmt.Sprintf("application/atom+xml;type=%s;charset=utf-8", type1)
	w.Header().Set("Content-Type", cont)
	res, err := xml.Marshal(data)
	if err != nil {
		r.Text(w, 500, fmt.Sprint("Failed to serialize xml: ", err))
	} else {
		w.Write(res)
	}
}

func entryFromRow(row interface{}) (*atom.XMLEntry, error) {
	var id, title, content string
	var orderId int
	var err error

	switch r := row.(type) {
	case *sql.Row:
		err = r.Scan(&id, &title, &content, &orderId)
	case *sql.Rows:
		err = r.Scan(&id, &title, &content, &orderId)
	}
	if err != nil {
		return nil, err
	}

	xmlTitle := atom.XMLTitle{Raw: title}
	xmlContent := atom.XMLEntryContent{Raw: content}
	entry := atom.XMLEntry{
		Id:      &id,
		Title:   &xmlTitle,
		Content: &xmlContent,
		IntId:   &orderId,
	}
	return &entry, nil
}

func addEntry(w http.ResponseWriter, req *http.Request) {
	entry, err := atom.DecodeEntry(req.Body)
	if err != nil {
		r.Text(w, 400, fmt.Sprint("could not parse xml: ", err))
		return
	}
	feedTitle := mux.Vars(req)["feed"]
	if _, err := insertEntry(entry, feedTitle); err != nil {
		r.Text(w, 500, fmt.Sprint("failed to save entry: ", err))
		return
	}
	r.XML(w, 201, entry)
}

func insertEntry(entry *atom.XMLEntry, feedTitle string) (*sql.Result, error) {
	_, err := findFeed(feedTitle)
	if err != nil {
		if err == sql.ErrNoRows {
			err = insertFeed(feedTitle)
			if err != nil {
				pqerr, ok := err.(*pq.Error)
				// gracefully handle race condition
				if !(ok && pqerr.Code.Name() == "unique_violation") {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}

	id := genId()
	titleType := "text"
	if entry.Title.Type != nil {
		titleType = *entry.Title.Type
	}
	contentType := "text"
	if entry.Content.Type != nil {
		contentType = *entry.Content.Type
	}
	result, err := db.Exec(
		`insert into atom_entry
		(id, feed_title, title, title_type, content, content_type)
		values ($1,$2,$3,$4,$5,$6)`,
		id, feedTitle, entry.Title.Raw, titleType,
		entry.Content.Raw, contentType,
	)
	return &result, err
}

func findFeed(title string) (*atom.XMLFeed, error) {
	row := db.QueryRow(`select id from atom_feed where title = $1`, title)
	var id string
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	feed := atom.XMLFeed{Id: &id}
	xmlTitle := atom.XMLTitle{Raw: title}
	feed.Title = &xmlTitle
	return &feed, nil
}

func insertFeed(title string) error {
	id := genId()
	_, err := db.Exec(
		`insert into atom_feed (id, title) values ($1, $2)`, id, title)
	return err
}

func genId() string {
	return fmt.Sprintf("urn:uuid:%s", uuid.NewV4().String())
}

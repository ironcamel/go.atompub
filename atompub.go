package atompub

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/ironcamel/go.atom"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/unrolled/render" // or "gopkg.in/unrolled/render.v1"
)

var r = render.New()
var db *sql.DB
var baseURL string = os.Getenv("GO_ATOMPUB_BASE_URL")

type AtomPub struct {
	BaseURL  string
	DSN      string
	Listener net.Listener
	Port     int
}

func (ap *AtomPub) Start() {
	var err error

	if ap.DSN == "" {
		ap.DSN = "postgres://localhost/atompub?sslmode=disable"
	}
	if db, err = sql.Open("postgres", ap.DSN); err != nil {
		log.Fatal("Could not open db: ", err)
	}
	_, err = db.Query("select 1")
	if err != nil {
		log.Fatal("Could not talk to db: ", err)
	}

	isCustomListener := true
	if ap.Listener == nil {
		isCustomListener = false
		if ap.Port == 0 {
			ap.Port = 8000
		}
		ap.Listener, _ = net.Listen("tcp", fmt.Sprint(":", ap.Port))
	}

	if ap.BaseURL == "" {
		ap.BaseURL = "http://localhost"
	}
	baseURL = ap.BaseURL

	router := mux.NewRouter()
	router.HandleFunc("/feeds/{feed}", getFeed).Methods("GET")
	router.HandleFunc("/feeds/{feed}", addEntry).Methods("POST")
	router.HandleFunc("/feeds/{feed}/entries/{entry}", getEntry).Methods("GET")
	router.HandleFunc("/status", getStatus).Methods("GET")
	if isCustomListener {
		// If a custom listener was passed in, we don't know the port or socket
		log.Println("AtomPub server starting ...")
	} else {
		log.Println("AtomPub server listening on port", ap.Port, "...")
	}
	http.Serve(ap.Listener, router)
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
	startAfter := req.FormValue("start-after")
	if startAfter == "" {
		startAfter = req.FormValue("start_after")
	}
	if err := populateFeed(feedPtr, startAfter); err != nil {
		r.Text(w, 500, fmt.Sprint("Failed to construct feed: ", err))
		return
	}

	numEntries := len(feedPtr.Entries)
	if numEntries > 0 {
		lastEntryId := feedPtr.Entries[numEntries-1].Id
		href := fmt.Sprintf("%s/feeds/%s?start-after=%s",
			baseURL, feedPtr.Title.Raw, *lastEntryId)
		rel := "next"
		nextLink := atom.XMLLink{Href: &href, Rel: &rel}
		feedPtr.Links = []atom.XMLLink{nextLink}
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

func getStatus(w http.ResponseWriter, req *http.Request) {
	r.Text(w, 200, "ok")
}

func findEntry(id string) (*atom.XMLEntry, error) {
	row := db.QueryRow(
		`select id, title, content, order_id, updated
		from atom_entry
		where id = $1`,
		id,
	)
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
		`select id, title, content, order_id, updated
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

func formatTime(t time.Time) *string {
	timeStr := t.Format(time.RFC3339)
	return &timeStr
}

func entryFromRow(row interface{}) (*atom.XMLEntry, error) {
	var id, title, content string
	var orderId int
	var updated time.Time
	var err error

	switch r := row.(type) {
	case *sql.Row:
		err = r.Scan(&id, &title, &content, &orderId, &updated)
	case *sql.Rows:
		err = r.Scan(&id, &title, &content, &orderId, &updated)
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
		Updated: formatTime(updated),
	}
	return &entry, nil
}

func insertEntry(entry *atom.XMLEntry, feedTitle string) (*sql.Result, error) {
	err := insertFeed(feedTitle)
	if err != nil {
		pqerr, ok := err.(*pq.Error)
		if !(ok && pqerr.Code.Name() == "unique_violation") {
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
	row := db.QueryRow(
		`select id, updated from atom_feed where title = $1`, title)
	var id string
	var updated time.Time
	if err := row.Scan(&id, &updated); err != nil {
		return nil, err
	}
	feed := atom.XMLFeed{Id: &id}
	xmlTitle := atom.XMLTitle{Raw: title}
	feed.Title = &xmlTitle
	feed.Updated = formatTime(updated)
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

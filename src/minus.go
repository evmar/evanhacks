package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

var feedsToFetch = []string{
	"108089472890519702166", // Consumer Surveys
	"109695781128233963799", // Andrew Jackson
}

type Fetch struct {
	Id   string
	Time time.Time
	Raw  []byte
	Feed []byte
}

func init() {
	http.HandleFunc("/feed/", frontPage)
	http.HandleFunc("/feed/cron", handleCron)
}

func getFetches(c appengine.Context) ([]*Fetch, error) {
	q := datastore.NewQuery("fetch").Order("-Time").Limit(100)
	i := q.Run(c)
	var fetches []*Fetch
	for {
		var f Fetch
		_, err := i.Next(&f)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		f.Raw = nil
		f.Feed = nil
		fetches = append(fetches, &f)
	}
	return fetches, nil
}

func frontPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/feed/" {
		handleFeed(w, r)
		return
	}

	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fetches, err := getFetches(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<pre>\n")
	for _, f := range fetches {
		fmt.Fprintf(w, "<a href='/feed/%s'>%s</a> %v\n", f.Id, f.Id, f.Time)
	}
}

func fetchGP(c appengine.Context, id string) (*Fetch, error) {
	url := fmt.Sprintf("https://www.googleapis.com/plus/v1/people/%s/activities/public?key=%s", id, privateApiKey)

	client := urlfetch.Client(c)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	raw := &bytes.Buffer{}
	io.Copy(raw, resp.Body)

	feed := &bytes.Buffer{}
	if err = transcode(raw, feed); err != nil {
		return nil, err
	}

	fetch := &Fetch{
		Id:   id,
		Time: time.Now(),
		Raw:  raw.Bytes(),
		Feed: feed.Bytes(),
	}
	_, err = datastore.Put(c, datastore.NewIncompleteKey(c, "fetch", nil), fetch)
	if err != nil {
		return nil, err
	}

	return fetch, nil
}

func handleCron(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	for _, id := range feedsToFetch {
		fetch, err := fetchGP(c, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "ok %q", fetch)
	}
}

func lookupFetch(c appengine.Context, id string) (*Fetch, error) {
	q := datastore.NewQuery("fetch").Filter("Id =", id).Order("-Time").Limit(1)
	i := q.Run(c)
	var f Fetch
	_, err := i.Next(&f)
	if err == datastore.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func handleFeed(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	prefix := "/feed/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	id := r.URL.Path[len(prefix):]
	f, err := lookupFetch(c, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if f == nil {
		http.Error(w, "no such feed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml")
	w.Header().Set("Expires", f.Time.Add(5*time.Minute).Format(time.RFC1123))
	w.Header().Set("Cache-Control", "public")
	http.ServeContent(w, r, "", f.Time, bytes.NewReader(f.Feed))
}

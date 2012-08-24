package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"time"
)

type GPFeed struct {
	Title string    `json:"title"`
	Id    string    `json:"id"`
	Items []*GPItem `json:"items"`
}
type GPActor struct {
	DisplayName string `json:"displayName"`
}
type GPItem struct {
	Title     string   `json:"title"`
	Id        string   `json:"id"`
	Actor     GPActor  `json:"actor"`
	Updated   string   `json:"updated"`
	Published string   `json:"published"`
	Object    GPObject `json:"object"`
}
type GPObject struct {
	Content string `json:"content"`
	Url     string `json:"url"`
}

type AtomFeed struct {
	XMLName xml.Name     `xml:"feed"`
	XMLNS   string       `xml:"xmlns,attr"`
	Id      string       `xml:"id"`
	Title   string       `xml:"title"`
	Updated string       `xml:"updated"`
	Entry   []*AtomEntry `xml:"entry"`
}
type AtomAuthor struct {
	Name string `xml:"name"`
}
type AtomEntry struct {
	Title     string      `xml:"title"`
	Id        string      `xml:"id"`
	Author    AtomAuthor  `xml:"author"`
	Updated   string      `xml:"updated"`
	Published string      `xml:"published"`
	Summary   AtomText    `xml:"summary"`
	Link      []*AtomLink `xml:"link"`
}
type AtomText struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}
type AtomLink struct {
	Href string `xml:"href,attr"`
}

func transcode(r io.Reader, w io.Writer) error {
	gpfeed := GPFeed{}
	if err := json.NewDecoder(r).Decode(&gpfeed); err != nil {
		return err
	}

	now := time.Now()
	atomfeed := AtomFeed{
		Title:   gpfeed.Title,
		Id:      gpfeed.Id,
		XMLNS:   "http://www.w3.org/2005/Atom",
		Updated: now.In(time.UTC).Format("2006-01-02T15:04:05Z"),
	}
	for _, item := range gpfeed.Items {
		link := AtomLink{
			Href: item.Object.Url,
		}
		entry := AtomEntry{
			Title:     item.Title,
			Id:        "tag:google.com,1970:" + item.Id,
			Author:    AtomAuthor{Name: item.Actor.DisplayName},
			Updated:   item.Updated,
			Published: item.Published,
			Summary:   AtomText{Type: "html", Text: item.Object.Content},
			Link:      []*AtomLink{&link},
		}
		atomfeed.Entry = append(atomfeed.Entry, &entry)
	}
	if err := xml.NewEncoder(w).Encode(&atomfeed); err != nil {
		return err
	}
	return nil
}

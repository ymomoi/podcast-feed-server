package main

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/f440/podcast-feed-server/config"
)

type RSS struct {
	XMLName        xml.Name `xml:"rss"`
	XMLXmlnsAtom   string   `xml:"xmlns:atom,attr"`
	XMLXmlnsItunes string   `xml:"xmlns:itunes,attr"`
	XMLVersion     string   `xml:"version,attr"`
	Channel        *Channel `xml:"channel,omitempty"`
}

type Channel struct {
	Title          string        `xml:"title,omitempty"`
	Description    string        `xml:"description,omitempty"`
	Link           string        `xml:"link,omitempty"`
	Language       string        `xml:"language,omitempty"`
	Copyright      string        `xml:"copyright,omitempty"`
	ChannelImage   *ChannelImage `xml:"image,omitempty"`
	Item           []*Item       `xml:"item,omitempty"`
	LastBuildDate  string        `xml:"lastBuildDate,omitempty"`
	AtomLink       *AtomLink     `xml:"atom:link,omitempty"`
	ItunesSubtitle string        `xml:"itunes:subtitle,omitempty"`
	ItunesAuthor   string        `xml:"itunes:author,omitempty"`
	ItunesSummary  string        `xml:"itunes:summary,omitempty"`
	ItunesKeywords string        `xml:"itunes:keywords,omitempty"`
	ItunesExplicit string        `xml:"itunes:explicit,omitempty"`
	ItunesOwner    *ItunesOwner  `xml:"itunes:owner,omitempty"`
	ItunesImage    *ItunesImage
}

type ItunesOwner struct {
	ItunesName  string `xml:"itunes:name,omitempty"`
	ItunesEmail string `xml:"itunes:mail,omitempty"`
}

type ItunesImage struct {
	XMLName xml.Name `xml:"itunes:image,omitempty"`
	Href    string   `xml:"href,attr"`
}

type ChannelImage struct {
	URL   string `xml:"url,omitempty"`
	Title string `xml:"title,omitempty"`
	Link  string `xml:"link,omitempty"`
}

type AtomLink struct {
	Href string `xml:"href,attr,omitempty"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"xml,attr,omitempty"`
}

type Item struct {
	Title                   string     `xml:"title,omitempty"`
	Description             string     `xml:"description,omitempty"`
	Guid                    string     `xml:"guid,omitempty"`
	PubDate                 string     `xml:"pubDate,omitempty"`
	Enclosure               *Enclosure `xml:"enclosure,omitempty"`
	ItunesAuthor            string     `xml:"itunes:author,omitempty"`
	ItunesSubtitle          string     `xml:"itunes:subtitle,omitempty"`
	ItunesSummary           string     `xml:"itunes:summary,omitempty"`
	ItunesDuration          string     `xml:"itunes:duration,omitempty"`
	ItunesExplicit          string     `xml:"itunes:explicit,omitempty"`
	ItunesOrder             string     `xml:"itunes:order,omitempty"`
	ItunesisClosedCaptioned string     `xml:"itunes:isClosedCaptioned,omitempty"`
}

type Enclosure struct {
	URL    string `xml:"url,attr,omitempty"`
	Type   string `xml:"type,attr,omitempty"`
	Length int64  `xml:"length,attr,omitempty"`
}

func EscapeURL(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func main() {
	config := config.Config{}
	if err := config.Load("config.toml"); err != nil {
		panic(err)
	}

	fs := http.FileServer(http.Dir(config.Server.FileRoot))
	http.Handle("/", fs)

	http.HandleFunc(config.Server.FeedPath, FeedHandler)

	http.ListenAndServe(config.Server.Listen, nil)
}

func FeedHandler(w http.ResponseWriter, r *http.Request) {
	config := config.Config{}
	if err := config.Load("config.toml"); err != nil {
		panic(err)
	}

	rss := RSS{
		XMLXmlnsAtom:   "http://www.w3.org/2005/Atom",
		XMLXmlnsItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		XMLVersion:     "2.0",
	}
	rss.Channel = &Channel{
		Title:       config.RSS.Title,
		Description: config.RSS.Description,
		Link:        config.RSS.URL,
		AtomLink: &AtomLink{
			Href: config.RSS.URL + config.Server.FeedPath,
			Rel:  "self",
			Type: "application/rss+xml",
		},
	}

	err := filepath.Walk(config.Server.FileRoot, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".mp3") {
			return nil
		}

		name := strings.Replace(info.Name(), ".mp3", "", 1)
		pubDate := info.ModTime().Format(time.RFC1123)
		url, err := EscapeURL(config.RSS.URL + strings.Replace(path, config.Server.FileRoot, "", 1))
		if err != nil {
			panic(err)
		}
		enclosure := Enclosure{URL: url, Type: "audio/mpeg", Length: info.Size()}
		item := Item{
			Title:     name,
			PubDate:   pubDate,
			Guid:      url,
			Enclosure: &enclosure,
		}

		rss.Channel.Item = append(rss.Channel.Item, &item)
		return nil
	})
	if err != nil {
		panic(err)
	}

	buf, _ := xml.MarshalIndent(rss, "", " ")

	w.Header().Set("Content-Type", "application/atom+xml")
	w.Write([]byte(xml.Header))
	w.Write(buf)
}

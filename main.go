package main

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Title       string
	Description string
	URL         string
	FeedURL     string `toml:"feed_url"`
	FileRoot    string `toml:"file_root"`
}

type RSS struct {
	XMLName        xml.Name `xml:"rss"`
	XMLXmlnsAtom   string   `xml:"xmlns:atom,attr"`
	XMLXmlnsItunes string   `xml:"xmlns:itunes,attr"`
	XMLVersion     string   `xml:"version,attr"`
	Channel        *Channel `xml:"channel,omitempty"`
}

type Channel struct {
	Title          string       `xml:"title,omitempty"`
	Description    string       `xml:"description,omitempty"`
	Link           string       `xml:"link,omitempty"`
	Language       string       `xml:"language,omitempty"`
	Copyright      string       `xml:"copyright,omitempty"`
	ChannelImage   ChannelImage `xml:"image,omitempty"`
	Item           []Item       `xml:"item,omitempty"`
	PubDate        string       `xml:"pubDate,omitempty"`
	LastBuildDate  string       `xml:"lastBuildDate,omitempty"`
	AtomLink       *AtomLink    `xml:"atom:link,omitempty"`
	ItunesSubtitle string       `xml:"itunes:subtitle,omitempty"`
	ItunesAuthor   string       `xml:"itunes:author,omitempty"`
	ItunesSummary  string       `xml:"itunes:summary,omitempty"`
	ItunesKeywords string       `xml:"itunes:keywords,omitempty"`
	ItunesExplicit string       `xml:"itunes:explicit,omitempty"`
	ItunesOwner    ItunesOwner  `xml:"itunes:owner,omitempty"`
	// ItunesImage
	// ItunesCategory
	// ItunesComplete
}

type ItunesOwner struct {
	ItunesName  string `xml:"itunes:name,omitempty"`
	ItunesEmail string `xml:"itunes:mail,omitempty"`
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
	Title                   string    `xml:"title,omitempty"`
	Description             string    `xml:"description,omitempty"`
	Guid                    string    `xml:"guid,omitempty"`
	PubDate                 string    `xml:"pubDate,omitempty"`
	Enclosure               Enclosure `xml:"enclosure,omitempty"`
	ItunesAuthor            string    `xml:"itunes:author,omitempty"`
	ItunesSubtitle          string    `xml:"itunes:subtitle,omitempty"`
	ItunesSummary           string    `xml:"itunes:summary,omitempty"`
	ItunesDuration          string    `xml:"itunes:duration,omitempty"`
	ItunesExplicit          string    `xml:"itunes:explicit,omitempty"`
	ItunesOrder             string    `xml:"itunes:order,omitempty"`
	ItunesisClosedCaptioned string    `xml:"itunes:isClosedCaptioned,omitempty"`
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
	now := time.Now().Format(time.RFC822)
	var config Config
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		panic(err)
	}

	rss := RSS{
		XMLXmlnsAtom:   "http://www.w3.org/2005/Atom",
		XMLXmlnsItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		XMLVersion:     "2.0",
	}
	rss.Channel = &Channel{
		Title:       config.Title,
		Description: config.Description,
		Link:        config.URL,
		PubDate:     now,
		AtomLink: &AtomLink{
			Href: config.FeedURL,
			Rel:  "self",
			Type: "application/rss+xml",
		},
	}

	os.Chdir(config.FileRoot)
	root := "."
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".mp3") {
			return nil
		}

		name := strings.Replace(info.Name(), ".mp3", "", -1)
		pubDate := info.ModTime().Format(time.RFC1123)
		url, err := EscapeURL(config.URL + path)
		if err != nil {
			panic(err)
		}
		enclosure := Enclosure{URL: url, Type: "audio/mpeg", Length: info.Size()}
		item := Item{
			Title:     name,
			PubDate:   pubDate,
			Guid:      url,
			Enclosure: enclosure,
		}

		rss.Channel.Item = append(rss.Channel.Item, item)
		return nil
	})
	if err != nil {
		panic(err)
	}

	buf, _ := xml.MarshalIndent(rss, "", " ")

	fmt.Println(xml.Header + string(buf))
}

package main

import (
	"bufio"
	"encoding/xml"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
	"github.com/dustin/go-humanize"
)

var funcs = template.FuncMap{
	"humanize_bytes": humanize.Bytes,
}

var views = template.Must(template.New("").Funcs(funcs).ParseGlob("views/*.html"))

func main() {
	log.SetHandler(logfmt.New(os.Stdout))
	addr := ":" + os.Getenv("PORT")
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/", index)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("error listening: %s", err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	views.ExecuteTemplate(w, "index.html", nil)
}

func submit(w http.ResponseWriter, r *http.Request) {
	file, hdr, err := r.FormFile("xml")
	if err != nil {
		log.WithError(err).Error("parsing form")
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	decoder := xml.NewDecoder(reader)

	var item EPGtv

	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}
		switch se := token.(type) {
		case xml.StartElement:
			decoder.DecodeElement(&item, &se)
		}
	}

	w.Header().Set("Content-Type", "text/html")
	err = views.ExecuteTemplate(w, "index.html", struct {
		Name  string
		Size  uint64
		Type  string
		EPGtv EPGtv
	}{
		Name:  hdr.Filename,
		Size:  uint64(hdr.Size),
		Type:  hdr.Header.Get("Content-Type"),
		EPGtv: item,
	})

	if err != nil {
		log.WithError(err).Error("parsing form")
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}

}

type EPGcatchup struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGcategory struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGcna_rating struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGcredits struct {
}

type EPGdesc struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGepisode_num struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGformat struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGicon struct {
	Attr_src string `xml:" src,attr"  json:",omitempty"`
}

type EPGprogramme struct {
	Attr_channel  string         `xml:" channel,attr"  json:",omitempty"`
	Attr_id       string         `xml:" id,attr"  json:",omitempty"`
	Start         customTime     `xml:" start,attr"  json:",omitempty"`
	Stop          customTime     `xml:" stop,attr"  json:",omitempty"`
	EPGcatchup    *EPGcatchup    `xml:" catchup,omitempty" json:"catchup,omitempty"`
	EPGcategory   *EPGcategory   `xml:" category,omitempty" json:"category,omitempty"`
	EPGcna_rating *EPGcna_rating `xml:" cna-rating,omitempty" json:"cna-rating,omitempty"`
	EPGcredits    *EPGcredits    `xml:" credits,omitempty" json:"credits,omitempty"`
	EPGdesc       *EPGdesc       `xml:" desc,omitempty" json:"desc,omitempty"`
	EPGformat     *EPGformat     `xml:" format,omitempty" json:"format,omitempty"`
	EPGicon       *EPGicon       `xml:" icon,omitempty" json:"icon,omitempty"`
	EPGreplay     *EPGreplay     `xml:" replay,omitempty" json:"replay,omitempty"`
	EPGseries     *EPGseries     `xml:" series,omitempty" json:"series,omitempty"`
	EPGtitle      *EPGtitle      `xml:" title,omitempty" json:"title,omitempty"`
}

type EPGreplay struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGroot struct {
	EPGtv *EPGtv `xml:" tv,omitempty" json:"tv,omitempty"`
}

type EPGseason_num struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGseries struct {
	EPGepisode_num *EPGepisode_num `xml:" episode-num,omitempty" json:"episode-num,omitempty"`
	EPGseason_num  *EPGseason_num  `xml:" season-num,omitempty" json:"season-num,omitempty"`
	EPGseries_id   *EPGseries_id   `xml:" series-id,omitempty" json:"series-id,omitempty"`
	EPGseries_name *EPGseries_name `xml:" series-name,omitempty" json:"series-name,omitempty"`
}

type EPGseries_id struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGseries_name struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGtitle struct {
	Text string `xml:",chardata" json:",omitempty"`
}

type EPGtv struct {
	EPGprogramme []*EPGprogramme `xml:" programme,omitempty" json:"programme,omitempty"`
}

type customTime struct {
	time.Time
}

func (c *customTime) UnmarshalXMLAttr(attr xml.Attr) error {
	const shortForm = "20060102150405"
	parse, err := time.Parse(shortForm, attr.Value)
	if err != nil {
		return err
	}
	*c = customTime{parse}
	return nil
}

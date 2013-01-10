package main

import (
	"exp/html"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	page       = "http://r-v-z.ch/index.php?id=64&group_ID=7072&team_ID=20160&nextPage=2"
	timeLayout = "02.01.06 15:04"
)

const (
	inStart = iota
	inTable
	inTR
)

const htmlTemplate = `
<!DOCTYPE html>
<!--[if lt IE 7]>      <html class="no-js lt-ie9 lt-ie8 lt-ie7"> <![endif]-->
<!--[if IE 7]>         <html class="no-js lt-ie9 lt-ie8"> <![endif]-->
<!--[if IE 8]>         <html class="no-js lt-ie9"> <![endif]-->
<!--[if gt IE 8]><!--> <html class="no-js"> <!--<![endif]-->
    <head>
        <meta charset="utf-8">
        <meta http-equiv="X-UA-Compatible" content="IE=edge" />
        <title>RAZ FAZ - Spielplan 2012/2013</title>
        <meta name="description" content="" />
        <meta name="viewport" content="width=device-width" />
        <meta name="format-detection" content="telephone=no" />
        <link rel="stylesheet" href="static/css/normalize.css" />
        <link rel="stylesheet" href="static/css/main.css" />
        <link href='http://fonts.googleapis.com/css?family=Oswald:700' rel='stylesheet' type='text/css' />
    </head>
    <body>
        <header>
          <h1>RAZ FAZ</h1>
          <h2>SAISON 2012/2013</h2> 
        </header>
        <section id="games"></section>
				{{range .}}
        <article class="game">
        <div class="info">
          <div class="venue">{{.Venue}}</div>
          <div class="date">{{.Date.Format "Mon 02.01.06"}}</div>
          <div class="time">{{.Date.Format "15:04"}}</div>
          <div class="opponent">{{.TeamOther}}</div>
        </div>
        <div class="result">{{.Result}}</div>
        </article>
				{{end}}
    </body>
</html>
`

type MatchEntry struct {
	Date      time.Time
	Venue     string
	TeamOther string
	Result    string
	Place     string
}

func main() {
	// Argument Parsing
	var port = flag.Int("port", 8080, "Listen port")
	var help = flag.Bool("help", false, "Show help.")
	flag.Parse()
	if *help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// HTTP Server
	// Static files
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	// Request handler
	http.HandleFunc("/", handleFunc)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening for request on address %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	// Retrieve web page
	resp, err := http.Get(page)
	if err != nil {
		log.Printf("Could not get page %s. Error: %v", page, err)
		return
	}
	defer resp.Body.Close()

	// HTML handling
	entries := parseHtml(&resp.Body)

	// Template handling
	tpl, err := template.New("razfaz").Parse(htmlTemplate)
	if err != nil {
		log.Printf("Could not parse template. Error %v", err)
		return
	}
	// Send it to client
	if err := tpl.Execute(w, entries); err != nil {
		log.Printf("Could not execute template. Error %v", err)
		return
	}
}

func parseHtml(body *io.ReadCloser) (entries []MatchEntry) {
	entries = make([]MatchEntry, 0, 10)
	z := html.NewTokenizer(*body)
	tableState := inStart
	done := false
	for {
		// All done.
		if done == true {
			return
		}

		// html end reached
		if z.Next() == html.ErrorToken {
			if z.Err() == io.EOF {
				log.Printf("End of file reached. This is suspicious.")
				return
			}
			log.Printf("Error while parsing: %v", z.Err())
			return
		}

		// tokenize & parse
		t := z.Token()
		switch tableState {
		case inStart:
			{
				if t.Type == html.StartTagToken && t.Data == "table" {
					for _, a := range t.Attr {
						if a.Key == "class" && a.Val == "tx_clicsvws_pi1_mainTableGroupResultsTable" {
							// We found our table with the data we are interested in. Go ahead an parse it.
							tableState = inTable
						}
					}
				}
			}
		case inTable:
			{
				if t.Type == html.EndTagToken && t.Data == "table" {
					// End of our table. All done.
					done = true
					break
				}

				if t.Type == html.StartTagToken && t.Data == "tr" {
					// We found a tr. Go ahead and parse all td's within.
					tableState = inTR
				}
			}
		case inTR:
			{
				var entry MatchEntry
				// Walk all td's
				tdCount := 0
				inTD := false
				for {
					z.Next()
					tt := z.Token()
					if tt.Type == html.EndTagToken && tt.Data == "tr" {
						// We are at the end of a tr. The entry should be parsed.
						entries = append(entries, entry)
						tableState = inTable
						break
					}

					// Check for beginning and end of a td tag.
					if tt.Type == html.StartTagToken && tt.Data == "td" {
						inTD = true
						tdCount += 1
					} else if tt.Type == html.EndTagToken && tt.Data == "td" {
						inTD = false
					}

					// We are only interested in text tokens.
					if inTD == true && tt.Type == html.TextToken {
						value := strings.TrimSpace(tt.Data)
						if value == "" {
							// Skip all space-like values
							continue
						}
						if value == "Datum" {
							// Ignore header row
							tableState = inTable
							break
						}
						// Match all columns in a row by td number
						switch tdCount {
						case 1:
							{
								// Date
								if v, err := time.Parse(timeLayout, value); err == nil {
									entry.Date = v
								}
							}
						case 4:
							{
								// Home team
								if value == "Raz Faz" {
									entry.Venue = "H"
								} else {
									entry.Venue = "A"
									entry.TeamOther = value
								}
							}
						case 6:
							{
								if value != "Raz Faz" {
									entry.TeamOther = value
								}
							}
						case 7:
							{
								// Result
								entry.Result = value
							}
						case 8:
							{
								// Place
								entry.Place = value
							}
						}
					}
				}
			}
		}
	}

	return
}

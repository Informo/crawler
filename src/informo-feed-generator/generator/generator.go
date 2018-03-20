package generator

import (
	"fmt"
	"net/http"
	"net/url"

	"common"
	"common/config"
	"common/database"

	// Using our own fork of feeds from now until our PR is accepted and merged
	// cf https://github.com/gorilla/feeds/pull/43
	"github.com/Informo/feeds"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Generator represents a RSS generator.
type Generator struct {
	db  *database.Database
	mux *mux.Router
	cfg config.FeedsConfig
}

// NewGenerator instantiate a new Generator.
func NewGenerator(db *database.Database, cfg config.FeedsConfig) *Generator {
	return &Generator{
		db:  db,
		mux: mux.NewRouter(),
		cfg: cfg,
	}
}

// SetupAndServe sets up the Generator's router and starts a web server that
// serves it at a given interface and port.
func (g *Generator) SetupAndServe() error {
	g.setup()
	listenAddr := fmt.Sprintf("%s:%d", g.cfg.Interface, g.cfg.Port)
	return http.ListenAndServe(listenAddr, g.mux)
}

// setup sets up the router for the Generator by defining a global function that
// will request the database for articles from a website given in the HTTP query's
// URL and generate a feed from them.
// If the handler function encounters an error, it will log send the string
// "Internal server error" to the requester, log the error's message and return,
// thus aborting the process.
func (g *Generator) setup() {
	// This string will be sent as a response to the request if any error happens
	// in order to avoid sending sensitive data contained in the error's message.
	var intSrvErr = "Internal server error"

	// Define a global route that will take the website's name as the only element
	// that can follow the initial "/".
	g.mux.HandleFunc("/{website}", func(w http.ResponseWriter, req *http.Request) {
		// Parse the variables (i.e. get the website's name from the request's URL)
		vars := mux.Vars(req)
		// Define a logger specificly for logging errors since we need to call it
		// from several places in this function.
		errLog := logrus.WithField("website", vars["website"])

		// Get the n latest articles for the requested website, n being a number
		// set in the configuration file.
		articles, err := g.db.RetrieveNLatestArticlesForWebsite(
			vars["website"], g.cfg.NbItems,
		)
		if err != nil {
			http.Error(w, intSrvErr, 500)
			errLog.Error(err)
			return
		}

		// If the slice of articles is empty, it means that there's not any article
		// from this website from the database.
		if len(articles) == 0 {
			http.Error(w, fmt.Sprintf("Unknown website %s", vars["website"]), 404)
			return
		}

		// Generate the gorilla/feeds representation of the feed we want to generate
		// with these articles.
		feed, err := g.getFeed(articles, vars["website"])
		if err != nil {
			http.Error(w, intSrvErr, 500)
			errLog.Error(err)
			return
		}

		// Convert this feed to string accordingly with the FeedType configuration
		// setting.
		feedStr, err := g.feedToString(feed)
		if err != nil {
			http.Error(w, intSrvErr, 500)
			errLog.Error(err)
			return
		}

		// Log each feed generation.
		logrus.WithFields(logrus.Fields{
			"website":        vars["website"],
			"content_length": len(feedStr),
			"feed_type":      g.cfg.Type,
			"nb_items":       g.cfg.NbItems,
		}).Info("Served feed")

		w.Header().Add("Content-Type", "text/xml;charset=utf-8")
		// Serve the feed.
		w.Write([]byte(feedStr))
	})
}

// getFeed generates a gorilla/feeds representation of a feed using the given articles
// and information about the site.
// Returns an error if there was an issue parsing a URL to get the website's base
// URL (scheme://host).
func (g *Generator) getFeed(articles []common.Article, website string) (*feeds.Feed, error) {
	// If this function is called, it means there's at least one article in the
	// slice. Because all articles supposedly come from the same website, we can
	// just take the first one to extract its scheme and host.
	u, err := url.Parse(articles[0].URL)
	if err != nil {
		return nil, err
	}

	// Allocate the feed structure.
	feed := &feeds.Feed{
		Title: website,
		Link:  &feeds.Link{Href: fmt.Sprintf("%s://%s", u.Scheme, u.Host)},
	}

	// Will serve as a buffer between the filling of the item and its appending
	// to the feed's slice of items. Declared here so we don't reallocate it in
	// each iteration of the loop.
	var i *feeds.Item
	for _, a := range articles {
		// Fill the item with the data we know are not nil.
		i = &feeds.Item{
			Title:   a.Title,
			Link:    &feeds.Link{Href: a.URL},
			Created: a.Date,
			Content: a.Content,
		}

		// Fill in the optional fields.
		if a.Description != nil {
			i.Description = *a.Description
		}
		if a.Author != nil {
			i.Author = &feeds.Author{Name: *a.Author}
		}

		// Append the item to the feed's slice of items.
		feed.Items = append(feed.Items, i)
	}

	return feed, nil
}

// feedToString generate the XML string from the given gorilla/feeds representation
// of the feed, accordingly with the FeedType specified in the configuration file,
// which should refer to either a RSS feed or an Atom feed.
// Returns an error if there was an issue generating the string.
func (g *Generator) feedToString(feed *feeds.Feed) (string, error) {
	switch g.cfg.Type {
	case config.FeedTypeRSS:
		return feed.ToRss()
	case config.FeedTypeAtom:
		return feed.ToAtom()
	}
	return "", nil
}

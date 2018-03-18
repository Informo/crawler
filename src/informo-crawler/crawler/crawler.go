// Copyright 2018 Informo core team <core@informo.network>
//
// Licensed under the GNU Affero General Public License, Version 3.0
// (the "License"); you may not use this file except in compliance with the
// License.
// You may obtain a copy of the License at
//
//     https://www.gnu.org/licenses/agpl-3.0.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crawler

import (
	"fmt"
	"time"

	"common/config"
	"common/database"

	"github.com/PuerkitoBio/gocrawl"
	"github.com/sirupsen/logrus"
)

// Crawler represents a website crawler, with its logger (a logrus logger with
// the "website" field prefilled), a gocrawl.Crawler instance and data on the
// website to crawl, along with the channels the extender will use to raise
// errors and request the crawl to be terminated.
type Crawler struct {
	Log     *logrus.Entry
	c       *gocrawl.Crawler
	website *config.Website
	errChan chan error
	endChan chan string
}

// NewCrawler takes configuration and database parameters, creates the channels
// the extender will use to raise errors and request the crawl to be terminated,
// uses all that data to instantiate an Extender. It then uses it and some
// configuration parameters to instantiate a Crawler.
// Returns an error if there was an issue instantiating the extender.
func NewCrawler(
	cfg config.CrawlerConfig, db *database.Database, website *config.Website,
) (*Crawler, error) {
	errChan := make(chan error)
	endChan := make(chan string)

	log := logrus.WithField("website", website.Identifier)
	// Instantiate the extender and the options.
	ext, err := NewExtender(db, website, log, errChan, endChan)
	if err != nil {
		return nil, err
	}
	opts := gocrawl.NewOptions(ext)

	// Fill the necessary options.
	opts.RobotUserAgent = cfg.RobotAgent
	opts.UserAgent = cfg.UserAgent
	opts.CrawlDelay = cfg.CrawlDelay * time.Second
	opts.MaxVisits = website.MaxVisits
	opts.LogFlags = gocrawl.LogInfo
	// opts.LogFlags = gocrawl.LogAll

	return &Crawler{
		Log:     log,
		c:       gocrawl.NewCrawlerWithOptions(opts),
		website: website,
		errChan: errChan,
		endChan: endChan,
	}, nil
}

// Run tells a crawler to start crawling, and watches the channels its extender
// will use to report error and to request the crawl to be aborted.
// When the crawl aborted, returns with the reason of its abortion.
func (c *Crawler) Run() (stopReason string) {
	var err error

	// We run the crawler in a goroutine. Since we also call this function in a
	// goroutine, and gocrawl's crawler starts another goroutine, it makes 3
	// intricated goroutines, which can seem to be a lot. However, because we
	// can't control gocrawl's goroutine, here's the lower level we can abort
	// a crawl from, which is why we run the crawler in another goroutine.
	go c.launchCrawler()

	// Watch on the crawler's channels for errors and abortion requests.
	for {
		select {
		case err = <-c.errChan:
			c.Log.Error(err)
		case stopReason = <-c.endChan:
			return stopReason
		}
	}
}

// launchCrawler runs the gocrawl's crawler instance. If the crawler terminates
// with an error, raises it except if it's gocrawl.ErrMaxVisits (because stopping
// after reaching the crawl limit is a normal and expected behaviour). In both
// cases it calls for abortion with the error's message as the reason.
// If there's no error (which means the crawler just finished working), requests
// an abortion without giving any reason.
func (c *Crawler) launchCrawler() {
	var err error

	if err = c.c.Run(c.website.StartPoint); err != nil && err != gocrawl.ErrMaxVisits {
		err = fmt.Errorf("Crawling failed: %v", err)
		c.errChan <- err
	}

	if err != nil {
		c.endChan <- err.Error()
	}

	c.endChan <- ""
}

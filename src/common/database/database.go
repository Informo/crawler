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

package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	// PostgreSQL driver
	_ "github.com/lib/pq"
)

// Database represents the crawler's database.
type Database struct {
	db       *sql.DB
	articles articlesStatements
}

// NewDatabase creates a new instance of the Database structure by opening a
// PostgreSQL database accessible using a given connexion configuration string,
// and preparing the different statements used.
// Returns an error if there was an issue opening the database or preparing the
// different statements.
func NewDatabase(connStr string) (database *Database, err error) {
	database = new(Database)

	if database.db, err = sql.Open("postgres", connStr); err != nil {
		return
	}
	if err = database.articles.prepare(database.db); err != nil {
		return
	}

	return
}

// SaveArticle saves an article into the database. description and author are
// optional, so pass them as a pointer that can be set to nil if the row field
// should be NULL.
// Returns an error if the insertion failed, or if the article's URL is invalid.
func (d *Database) SaveArticle(
	website string, articleURL *url.URL, title string,
	description *string, content string, author *string, date time.Time,
) error {
	// Check the article's URL.
	if articleURL.Scheme != "http" && articleURL.Scheme != "https" {
		return fmt.Errorf("Unsupported protocol scheme for provided URL: %s", articleURL.Scheme)
	}

	// Perform the insertion.
	return d.articles.insertArticle(
		website, articleURL.String(), title, description, content, author, date,
	)
}

// RetrieveArticleURLsForWebsite retrieves from the database all articles that
// were published on a given website.
// Returns an error if the retrieval failed.
func (d *Database) RetrieveArticleURLsForWebsite(website string) (map[string]bool, error) {
	return d.articles.selectArticlesURLsForWebsite(website)
}

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
	"time"
)

// Schema of the articles table.
const articlesSchema = `
-- Store articles found while crawling
CREATE TABLE IF NOT EXISTS articles (
	-- Website the article is from
	website TEXT NOT NULL,
	-- Article's URL
	url TEXT NOT NULL PRIMARY KEY,
	-- Article's title
	title TEXT NOT NULL,
	-- Article's description. Can be NULL.
	description TEXT,
	-- Article's content
	content TEXT NOT NULL,
	-- Article's author. Can be NULL.
	author TEXT,
	-- Article's date
	date DATE NOT NULL
);
`

// Retrieve URLs of all articles filtered by the website they were posted on.
const selectArticlesURLsForWebsiteSQL = `
	SELECT url FROM articles WHERE website = $1
`

// Insert a new article in the database.
const insertArticleSQL = `
	INSERT INTO articles (website, url, title, description, content, author, date)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`

type articlesStatements struct {
	selectArticlesURLsForWebsiteStmt *sql.Stmt
	insertArticleStmt                *sql.Stmt
}

// Create the table if it doesn't exist and prepare the SQL statements.
func (a *articlesStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(articlesSchema)
	if err != nil {
		return
	}
	if a.selectArticlesURLsForWebsiteStmt, err = db.Prepare(selectArticlesURLsForWebsiteSQL); err != nil {
		return
	}
	if a.insertArticleStmt, err = db.Prepare(insertArticleSQL); err != nil {
		return
	}
	return
}

// insertArticle inserts an article into the database. description and author
// are optional, so pass them as a pointer that can be set to nil if the row field
// should be NULL.
// Returns an error if there was an issue inserting the article.
func (a *articlesStatements) insertArticle(
	website string, url string, title string, description *string,
	content string, author *string, date time.Time,
) (err error) {
	var descriptionNullable, authorNullable sql.NullString

	// Optional fields.
	descriptionNullable.Valid = description != nil
	if descriptionNullable.Valid {
		descriptionNullable.String = *description
	}
	authorNullable.Valid = author != nil
	if authorNullable.Valid {
		authorNullable.String = *author
	}

	// Run the insertion.
	_, err = a.insertArticleStmt.Exec(
		website, url, title, descriptionNullable, content, authorNullable, date,
	)

	return
}

// selectArticlesURLsForWebsite returns the URLs of all articles published on
// a given website.
// Returns an error if there was an issue requesting the URLs from the database,
// or extracting the data from the rows.
func (a *articlesStatements) selectArticlesURLsForWebsite(website string) (urls map[string]bool, err error) {
	rows, err := a.selectArticlesURLsForWebsiteStmt.Query(website)
	if err != nil {
		return
	}

	// Using a map instead of an array here because we're going to store a lot
	// of URLs (hundreds, thousands, or even more) that we'll need to access
	// very frequently, and benchmarks show that retrieval is faster on maps
	// than in arrays.
	urls = make(map[string]bool)

	// Retrieve the URLs and save them in the map.
	var url string
	for rows.Next() {
		if err = rows.Scan(&url); err != nil {
			return
		}

		urls[url] = true
	}

	return
}

package common

import (
	"time"
)

// Article describes the representation of a news item.
type Article struct {
	URL         string
	Title       string
	Description *string
	Content     string
	Author      *string
	Date        time.Time
}

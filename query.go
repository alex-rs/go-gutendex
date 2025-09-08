package gutendex

import "net/url"

// Query describes filters for listing books.
type Query struct {
	Author   string
	Title    string
	Topic    string
	Language string
	MIME     string
}

// Values converts the query into URL values compatible with Gutendex.
func (q Query) Values() url.Values {
	v := url.Values{}
	if q.Author != "" && q.Title != "" {
		v.Set("search", q.Author+" "+q.Title)
	} else {
		if q.Author != "" {
			v.Set("author", q.Author)
		}
		if q.Title != "" {
			v.Set("title", q.Title)
		}
	}
	if q.Topic != "" {
		v.Set("topic", q.Topic)
	}
	if q.Language != "" {
		v.Set("languages", q.Language)
	}
	if q.MIME != "" {
		v.Set("mime_type", q.MIME)
	}
	return v
}

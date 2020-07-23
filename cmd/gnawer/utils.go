package main

import (
	"net/url"

	"github.com/jaytaylor/html2text"
)

func isValidURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	return true
}

func htmlToText(s string) string {
	s, _ = html2text.FromString(s, html2text.Options{PrettyTables: true, OmitLinks: true})
	return s
}

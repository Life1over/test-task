package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

func collectArticle(t *Task, url string) *Article {
	res, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil
	}

	title, _ := doc.Find(t.Title).Html()
	title = htmlToText(title)
	content, _ := doc.Find(t.Content).Html()
	content = htmlToText(content)
	if title != "" && content != "" {
		return &Article{URL: url, Title: title, Content: content}
	}
	return nil
}

func collectHTMLArticles(t *Task) ([]*Article, error) {
	res, err := http.Get(t.URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var articles []*Article
	doc.Find(t.Link).Each(func(i int, s *goquery.Selection) {
		if url, b := s.Attr("href"); b {
			if !isValidURL(url) {
				url = path.Join(t.Link, url)
				if !isValidURL(url) {
					return
				}
			}
			a := collectArticle(t, url)
			if a != nil {
				articles = append(articles, a)
			}
		}
	})
	return articles, nil
}

func collectRSSArticles(t *Task) ([]*Article, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(t.URL)
	if err != nil {
		return nil, err
	}
	var articles []*Article
	for _, f := range feed.Items {
		if f.Content == "" {
			f.Content = f.Description
		}
		articles = append(articles, &Article{URL: f.Link, Title: f.Title, Content: f.Content})
	}
	return articles, nil
}

func collectArticles(t *Task) ([]*Article, error) {
	if t.Kind == "RSS" {
		return collectRSSArticles(t)
	}
	return collectHTMLArticles(t)
}

func parseAdd(args []string) (*Task, error) {
	var t Task
	f := flag.NewFlagSet("add", flag.ExitOnError)
	f.StringVar(&t.Name, "n", "", "Unique name")
	f.StringVar(&t.Kind, "k", "HTML", "Kind: HTML|RSS. Default is HTML")
	f.StringVar(&t.URL, "u", "", "URL")
	f.StringVar(&t.Link, "l", "", "Link query. Example: a.item__link")
	f.StringVar(&t.Title, "t", "", "Title query. Example: div.article__header")
	f.StringVar(&t.Content, "c", "", "Content query. Example: div.article__text")
	if err := f.Parse(args); err != nil {
		return nil, err
	}
	t.Kind = strings.ToUpper(t.Kind)
	if t.Kind == "HTML" && (t.Name == "" || t.URL == "" || t.Link == "" || t.Title == "" || t.Content == "") {
		return nil, errors.New("Name, URL, Link query, title query and content query should be specified.")
	} else if t.Kind == "RSS" && (t.Name == "" || t.URL == "") {
		return nil, errors.New("Name and URL should be specified.")
	}
	if !isValidURL(t.URL) {
		return nil, errors.New("Invalid URL: " + t.URL)
	}
	return &t, nil
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("gnawer add [OPTION]...")
	fmt.Println("Try 'gnawer add --help' for more information.")
	fmt.Println(`gnawer news [search this text]`)
	fmt.Println("gnawer tasks")
	fmt.Println("gnawer update")
	os.Exit(1)
}

// Нужно все рефакторить, сложно для восприятия
func main() {
	storage, err := NewStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "add":
		t, err := parseAdd(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		if err := storage.AddTask(t); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Task \"%s\" added.\n", t.Name)
	case "tasks":
		tt, err := storage.ListTasks()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Tasks:")
		for _, t := range tt {
			fmt.Printf("%v\n", *t)
		}
	case "update":
		tt, err := storage.ListTasks()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Loaded news:")
		for _, t := range tt {
			aa, _ := collectArticles(t)
			for _, a := range aa {
				fmt.Printf("Title: %s; Link: %s\n", a.Title, a.URL)
				storage.AddArticle(a)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	case "news":
		fmt.Println("News:")
		if len(os.Args) >= 3 {
			fmt.Println("Search: " + strings.Join(os.Args[2:], " "))
		}
		aa, err := storage.ListArticles(os.Args[2:]...)
		if err != nil {
			log.Fatal(err)
		}
		for _, a := range aa {
			fmt.Printf("Title: %s\n%s\n\n", a.Title, a.Content)
		}
	default:
		usage()
	}
}

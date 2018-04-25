package main

import (
	"database/sql"
	"os"

	_ "github.com/cznic/ql/driver"
)

type Task struct {
	Name    string
	Kind    string
	URL     string
	Link    string
	Title   string
	Content string
}

type Article struct {
	URL     string
	Title   string
	Content string
}

type Storage struct {
	db *sql.DB
}

func (s *Storage) Close() error {
	if s == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Storage) updateTask(t *Task) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE Tasks SET Kind=$2, URL=$3, Link=$4, Title=$5, Content=$6
		WHERE Name==$1`, t.Name, t.Kind, t.URL, t.Link, t.Title, t.Content)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}
	return false, nil
}

func (s *Storage) addTask(t *Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`INSERT INTO Tasks 
		(Name, Kind, URL, Link, Title, Content) 
		VALUES($1, $2, $3, $4, $5, $6)`, t.Name, t.Kind, t.URL, t.Link, t.Title, t.Content); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) AddTask(t *Task) error {
	if b, err := s.updateTask(t); err != nil {
		return err
	} else if b {
		return nil
	}
	return s.addTask(t)
}

func (s *Storage) ListTasks() ([]*Task, error) {
	rows, err := s.db.Query("SELECT Name, Kind, URL, Link, Title, Content FROM Tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tt := []*Task{}
	for rows.Next() {
		t := &Task{}
		if err := rows.Scan(&t.Name, &t.Kind, &t.URL, &t.Link, &t.Title, &t.Content); err != nil {
			return nil, err
		}
		tt = append(tt, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tt, nil
}

func (s *Storage) updateArticle(a *Article) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE News SET Title=$2, Content=$3 WHERE URL==$1`, a.URL, a.Title, a.Content)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}
	return false, nil
}

func (s *Storage) addArticle(a *Article) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`INSERT INTO News (URL, Title, Content) 
		VALUES($1, $2, $3)`, a.URL, a.Title, a.Content); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) AddArticle(a *Article) error {
	if b, err := s.updateArticle(a); err != nil {
		return err
	} else if b {
		return nil
	}
	return s.addArticle(a)
}

func (s *Storage) ListArticles(contains ...string) ([]*Article, error) {
	var rows *sql.Rows
	var err error
	if len(contains) == 0 {
		rows, err = s.db.Query("SELECT URL, Title, Content FROM News")
	} else {
		r := "" //".*"
		for _, c := range contains {
			r += ".*" + c
		}
		r += ".*"
		rows, err = s.db.Query("SELECT URL, Title, Content FROM News WHERE Title LIKE $1", r)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aa := []*Article{}
	for rows.Next() {
		a := &Article{}
		if err := rows.Scan(&a.URL, &a.Title, &a.Content); err != nil {
			return nil, err
		}
		aa = append(aa, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return aa, nil
}

// TODO
// type Tasks struct {
// 	s *Storage
// }

// TODO
// type Articles struct {
// 	s *Storage
// }

func NewStorage() (Storage, error) {
	if err := os.Mkdir("db", os.ModePerm); err != nil {
		return Storage{db: nil}, err
	}
	db, err := sql.Open("ql", "db/news.ql")
	if err != nil {
		return Storage{db: nil}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return Storage{db: nil}, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS Tasks (
		Name string NOT NULL, 
		URL string NOT NULL, 
		Kind string Kind IN ("RSS", "HTML"), 
		Link string, 
		Scope string, 
		Title string NOT NULL, 
		Content string NOT NULL );
		CREATE UNIQUE INDEX IF NOT EXISTS TaskName ON Tasks (Name);`); err != nil {
		return Storage{db: nil}, err
	}
	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS News (
		URL string NOT NULL,
		Date time,
		Title string NOT NULL,
		Content string NOT NULL);
		CREATE UNIQUE INDEX IF NOT EXISTS NewsURL ON News (URL)`); err != nil {
		return Storage{db: nil}, err
	}
	if err := tx.Commit(); err != nil {
		return Storage{db: nil}, err
	}
	return Storage{db: db}, nil
}

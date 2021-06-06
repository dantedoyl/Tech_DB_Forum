package models

import (
	"encoding/json"
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Status struct {
	User   int `json:"user"`
	Forum  int `json:"forum"`
	Thread int `json:"thread"`
	Post   int `json:"post"`
}

type User struct {
	Nickname string `json:"nickname"`
	FullName string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
}

type Forum struct {
	ID      int    `json:"-"`
	Title   string `json:"title"`
	User    string `json:"user"`
	Slug    string `json:"slug"`
	Posts   int    `json:"posts"`
	Threads int    `json:"threads"`
}

type Thread struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Author  string    `json:"author"`
	Forum   string    `json:"forum"`
	Message string    `json:"message"`
	Votes   int       `json:"votes"`
	Slug    string    `json:"slug"`
	Created time.Time `json:"created"`
}

type ThreadWithoutSlug struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Author  string    `json:"author"`
	Forum   string    `json:"forum"`
	Message string    `json:"message"`
	Votes   int       `json:"votes"`
	Slug    string    `json:"-"`
	Created time.Time `json:"created"`
}

func DeleteSlug(thread *Thread) *ThreadWithoutSlug{
	return &ThreadWithoutSlug{
		ID:       thread.ID,
		Title:    thread.Title,
		Author:   thread.Author,
		Forum:    thread.Forum,
		Message:  thread.Message,
		Votes:    thread.Votes,
		Slug:     thread.Slug,
		Created:  thread.Created,
	}
}

type Post struct {
	ID       	int          `json:"id"`
	Author   	string       `json:"author"`
	Created  	time.Time    `json:"created"`
	Forum    	string       `json:"forum"`
	IsEdited 	bool         `json:"isEdited"`
	Message  	string       `json:"message"`
	Parent   	int64    	 `json:"parent"`
	Thread   	int          `json:"thread,"`
	Route    pgtype.Int8Array `json:"-"`
}

type PostUpdate struct {
	ID       	int       	`json:"-"`
	Message 	string 		`json:"message"`
}

type PostInfo struct {
	Author 		*User       `json:"author"`
	Forum  		*Forum      `json:"forum"`
	Post   		*Post       `json:"post"`
	Thread 		interface{} `json:"thread"`
}

type Vote struct {
	Nickname 	string 		`json:"nickname"`
	Voice    	int    		`json:"voice"`
	Thread   	int    		`json:"-"`
}

type Error struct {
	Code 		int 		`json:"-"`
	Message 	string 		`json:"message"`
}

func ErrorToJSON(error string) []byte{
	bytes, err := json.Marshal(Error{Message: error})
	if err != nil {
		return []byte("")
	}
	return bytes
}

type Params struct {
	Limit 		int    		`json:"limit"`
	Since 		string 		`json:"since"`
	Desc  		bool   		`json:"desc"`
	Sort 		string 		`json:"sort"`
}

type Related struct {
	IsUser 		bool
	IsForum 	bool
	IsThread 	bool
}
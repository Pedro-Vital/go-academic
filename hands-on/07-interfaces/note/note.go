package note

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Note struct {
	Title string `json:"title"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// For json content, it would be more common to have lowercase keys in the json keys,
// so we can use tags to replace the default json keys, which are the fields names
// Tags are just metadata
// The json package will look for metadata like `json:""`

func (note Note) Display() {
	fmt.Printf("Your note titled %v has the following content:\n\n%v\n\n", note.Title, note.Content)
}

func (note Note) Save() error {
	fileName := strings.ReplaceAll(note.Title, " ", "_")
	fileName = strings.ToLower(fileName) + ".json"

	json, err := json.Marshal(note)
	// that is a function that converts data to JSON
	// it works with various kinds of data, including structs
	// the fields of the structs must be public

	if err != nil {
		return err
	}

	return os.WriteFile(fileName, json, 0644)
	// I want to have one file per note and the file name should depend on the note title.
	// Here, I want to create a JSON file that contains data in that JSON format.
	// 0644 is read and edit permission to the owner of the file and only read permission
	// to any other user.
	// We return it because WriteFile() returns error

}

func New(title, content string) (Note, error) {
	if title == "" || content == "" {
		return Note{}, errors.New("title and content cannot be empty")
	}
	
	return Note{
		Title: title,
		Content: content,
		CreatedAt: time.Now(),
	}, nil
}

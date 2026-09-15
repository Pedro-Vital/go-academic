package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"structsproject/note"
)

func main() {
	title, content := getNoteData()

	userNote, err := note.New(title, content)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	userNote.Display()
	err = userNote.Save()

	if err != nil {
		fmt.Println("Saving the note failed.")
		return
	}

	fmt.Println("Saving the note succeeded")
}

func getNoteData() (string, string) {
	title := getUserInput("Note title:")

	content := getUserInput("Note content:")

	return title, content
}

func getUserInput(prompt string) string {
	fmt.Printf("%v ", prompt)

	reader := bufio.NewReader(os.Stdin)
	// The constructor is used to create a new reader that reads from the standard input

	text, err := reader.ReadString('\n')
	// We use the reader to read the string that was entered by the user.
	// ReadString wants to know at which byte it should stop reading,
	// '\n' in that case.
	// to specify such a value, we need single quotes.
	// this single *Unicode code point* value is a special value in Go called rune.

	if err != nil {
		return ""
	}

	// We'll still have the \n in the string, so we need to remove it:
	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimSuffix(text, "\r") // Sometimes there is this special character too

	return text
}

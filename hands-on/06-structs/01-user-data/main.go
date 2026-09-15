package main

import (
	"fmt"
	"userdata/user"
)

func main() {
	userFirstName := getUserData("Please enter your first name: ")
	userLastName := getUserData("Please enter your last name: ")
	userBirthdate := getUserData("Please enter your birthdate (MM/DD/YYYY): ")

	var appUser *user.User

	appUser, err := user.New(userFirstName, userLastName, userBirthdate)

	if err != nil {
		fmt.Println(err)
		return
	}

	admin := user.NewAdmin("pedro@example.com", "test123")

	// User methods are accessible by Admin due to promotion
	admin.OutputUserDetails()
	admin.ClearUserName()
	admin.OutputUserDetails()
	// admin is a value and ClearUserName() is a pointer receiver.
	// Go treats this as (&admin).ClearUserName().
	// Go can implicitly take the address of an addressable value when calling a method with a pointer receiver.

	appUser.OutputUserDetails()
	// appuser is a pointer and OutputUserDetails() is a value receiver.
	// Go treats this as (*appUser).OutputUserDetails(). No need to dereference.
	appUser.ClearUserName()
	appUser.OutputUserDetails()
}

func getUserData(promptText string) string {
	fmt.Print(promptText)
	var value string
	fmt.Scanln(&value)
	return value
}

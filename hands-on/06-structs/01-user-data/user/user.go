package user

import (
	"errors"
	"fmt"
	"time"
)

// Declaration
type User struct { // Uppercase to export the User identifier (use it in other packages)
	firstName string // lowercase fields to make them accessible only within this package
	lastName  string
	birthdate string
	createdAt time.Time
}

// User is a value type, their instances are values

type Admin struct {
	email    string
	password string
	User     // Struct embedding
	// The fields and methods of User will be directly accessible by Admin due to the promotion.
}

// Method
// It is a value receiver. It will not modify the instance unless you reassign during the call.
func (u User) OutputUserDetails() {
	fmt.Println(u.firstName, u.lastName, u.birthdate)
}

/*
We could seamlessly override the OutputUserDetails() method for the admin.
Even though they have the same method name, this is allowed because methods are identified by:
receiver type + method name

func (a Admin) OutputUserDetails() {
	fmt.Println(a.firstName, a.lastName, a.birthdate, a.email)
}

admin.OutputUserDetails()      // calls Admin method
admin.User.OutputUserDetails() // calls User method
*/

// Mutation method
// It is a pointer receiver. It can be used to modify the instance of the struct
func (u *User) ClearUserName() {
	// There is no need to dereference using structs:
	// (*u).firstName = ""
	// (*u).lastName = ""
	// Go performs the dereference automatically:
	u.firstName = ""
	u.lastName = ""
}

// Constructor
func NewAdmin(email, password string) Admin {
	return Admin{ // Creating struct instance
		email:    email,
		password: password,
		User: User{ // The field name of an embedded struct is implicitly the type name
			firstName: "ADMIN",
			lastName:  "ADMIN",
			birthdate: "---",
			createdAt: time.Now(),
		},
	}
}

// Constructor
func New(firstName, lastName, birthdate string) (*User, error) {
	// Validation
	if firstName == "" || lastName == "" || birthdate == "" {
		return nil, errors.New("Missing input.")
	}

	return &User{
		// In this example, it's not strictly necessary to return a pointer,
		// but it's a convention, the idiomatic choice and scales well if User becomes more complex.
		// Returning a value copies this whole struct whenever it's returned or passed by value.
		// Returning a pointer avoids those copies.
		firstName: firstName,
		lastName:  lastName,
		birthdate: birthdate,
		createdAt: time.Now(),
	}, nil
}

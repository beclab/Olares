package util

import (
	"golang.org/x/crypto/bcrypt"
)

func Bcrypt(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
	// fmt.Println("Hashed Password:", string(hashedPassword))

	// err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	// if err != nil {
	// 	fmt.Println("Password is incorrect")
	// } else {
	// 	fmt.Println("Password is correct")
	// }

	// return "", nil
}

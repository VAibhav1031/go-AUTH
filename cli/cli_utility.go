package cli

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/chzyer/readline"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

/*Pre Login Functions */
type App struct {
	rl      *readline.Instance
	db      *sql.DB
	session Session
}

func (a *App) handleLogin() {
	a.rl.SetPrompt("Username: ")
	username, _ := a.rl.Readline()

	passwordBytes, _ := a.rl.ReadPassword("Password")
	password := string(passwordBytes)

	var storedPassword string
	// check the password
	getPasswordSQL := `SELECT password FROM users where username = ?`
	err := a.db.QueryRow(getPasswordSQL, username).Scan(&storedPassword)
	if err == sql.ErrNoRows {
		log.Println("User Not Found")
	} else if err != nil {
		log.Fatalf("Database error :%v", err)
	}

	if !checkPasswordHash(password, storedPassword) {
		log.Fatalf("Incorrect Password")
		fmt.Println("Incorrect Password!!")
	}

}

func (a *App) handleRegister() {

	a.rl.SetPrompt("Username: ")
	username, _ := a.rl.Readline()

	var confirmPassword string
	for {
		passwordBytes, _ := a.rl.ReadPassword("Password: ")
		password := string(passwordBytes)
		confirmPasswordBytes, _ := a.rl.ReadPassword("Confirm Password: ")
		confirmPassword := string(confirmPasswordBytes)

		if password != confirmPassword {
			continue
		} else {
			break
		}
	}
	hashedPassword, err := hashPassword(confirmPassword)
	if err != nil {

	}

	// create the record
	// fields must  userId, username, hashedpassword , failedAttempts , TimeBlocked
	insertSQL := `INSERT INTO users (username, password) VALUES (?, ?);`
	result, err := a.db.Exec(insertSQL, username, hashedPassword)
	if err != nil {
		log.Printf("Insert Failed for the Register Handle (likely duplicate Password or Username): %v", err)

	} else {
		lastID, _ := result.LastInsertId()
		fmt.Printf("Registered user with Id %d\n", lastID)

	}
	// populate the session here i would say soo that would be the helpful thing
}

func (a *App) handleHelp()

/* Post Login Functions */

func (a *App) handleWhoami()
func (a *App) handleLogout()

func (a *App) handleEnable2Fa()

func (a *App) handlePostHelp()

func (a *App) handleDisable2Fa()

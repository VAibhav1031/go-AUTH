package cli

import (
	"fmt"
	"time"

	"github.com/chzyer/readline"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const globalTryLimit = 3

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

type User struct {
	ID             int        `db:"id"`
	Username       string     `db:"username"`
	StoredPassword string     `db:"password"`
	CreatedAt      time.Time  `db:"created_at"`
	MFAEnabled     bool       `db:"mfa_enabled"`
	MFASecret      string     `db:"mfa_secret"`
	LastLogin      *time.Time `db:"last_login"`
	Attempts       int        `db:"attempts"`
	BlockedTime    time.Time  `db:"blocked_time"` // Safe handling for NULL times

}

type App struct {
	rl      *readline.Instance
	db      *sqlx.DB
	session *Session
}

// OTher Utility Functions
func handlePostHelp() {
	fmt.Println(`Available commands:
	whoami       Show current user details
	enable-2fa   Enable TOTP-based MFA
	disable-2fa  Disable MFA
	logout       End session
	help         Show this help message`)
}

func handleHelp() {

	fmt.Println(`Available commands:
	register   Create a new user
	login      Login with username/password (+ TOTP if enabled)
	help       Show this help message
	exit       Quit the program`)
}

func printUserDetails(u *User, s *Session) {

	fmt.Println("── Logged in ──")
	fmt.Printf("Username:      %s\n", u.Username)
	fmt.Printf("Registered on: %s\n", u.CreatedAt.Format("2006-01-02 15:04:05"))

	mfaStatus := "disabled"
	if u.MFAEnabled {
		mfaStatus = "enabled"
	}
	fmt.Printf("MFA status:    %s\n", mfaStatus)

	fmt.Printf("Session expires at: %s\n", s.ExpiresAt.Format("2006-01-02 15:04:05"))

	if u.LastLogin != nil {
		fmt.Printf("Last login:    %s\n", u.LastLogin.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Last login:    (first login)")
	}
}
func banner() {
	fmt.Println(`
 _______  _______         _______          _________                 _______  _       _________
(  ____ \(  ___  )       (  ___  )|\     /|\__   __/|\     /|       (  ____ \( \      \__   __/
| (    \/| (   ) |       | (   ) || )   ( |   ) (   | )   ( |       | (    \/| (         ) (   
| |      | |   | | _____ | (___) || |   | |   | |   | (___) | _____ | |      | |         | |   
| | ____ | |   | |(_____)|  ___  || |   | |   | |   |  ___  |(_____)| |      | |         | |   
| | \_  )| |   | |       | (   ) || |   | |   | |   | (   ) |       | |      | |         | |   
| (___) || (___) |       | )   ( || (___) |   | |   | )   ( |       | (____/\| (____/\___) (___
(_______)(_______)       |/     \|(_______)   )_(   |/     \|       (_______/(_______/\_______/`)
}

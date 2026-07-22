package cli

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/chzyer/readline"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const globaTryLimit = 3

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
	LastLogin      *time.Time `db:"last_login"`
	Attempts       int        `db:"attempts"`
	BlockedTime    time.Time  `db:"blocked_time"` // Safe handling for NULL times

}

type App struct {
	rl      *readline.Instance
	db      *sqlx.DB
	session *Session
}

/*Pre Login Methods */

func (a *App) handleLogin() {
	a.rl.SetPrompt("Username: ")
	username, _ := a.rl.Readline()

	passwordBytes, _ := a.rl.ReadPassword("Password")
	password := string(passwordBytes)

	userDetailSQL := `SELECT  * FROM users where username = ?;`

	var singleUser User
	err := a.db.Get(&singleUser, userDetailSQL, username)
	if err == sql.ErrNoRows {
		slog.Error("User Not Found")
		fmt.Println("User", username, "Not registered!!")
		return // user doesnt exist so exit the handler
	} else if err != nil {
		slog.Error("Database error :%v", err)
		fmt.Println("AUTH FAILURE")
		return
	}

	// Check for the blocked time passing
	if time.Now().Before(singleUser.BlockedTime) {
		slog.Info("User", singleUser.Username, "Still blocked!!")
		fmt.Println("Lockout for Many Incorrect Login Attempts!!")
		return
	}

	attempts := singleUser.Attempts
	// password comparison
	if !checkPasswordHash(password, singleUser.StoredPassword) {
		slog.Error("Incorrect Password")
		fmt.Println("Incorrect Password!!")

		// check for the attempt limit
		if attempts+1 >= globaTryLimit {
			// block: reset attempts to 0 once blocked, set the lockout window
			blockedUntil := time.Now().Add(15 * time.Minute)
			blockSQL := `UPDATE users SET attempts = ?, blocked_time = ? WHERE username = ?;`
			_, err := a.db.Exec(blockSQL, attempts+1, blockedUntil, username)
			if err != nil {
				slog.Error("Update failed", "error", err)
				fmt.Println("AUTH FAILURE")
			}
		} else {
			increaseSQL := `UPDATE users SET attempts = ? WHERE username = ?;`
			_, err := a.db.Exec(increaseSQL, attempts+1, username)
			if err != nil {
				slog.Error("Update failed", "error", err)
				fmt.Println("AUTH FAILURE")
			}
		}
		return
	} else {
		// successful login: reset attempts, clear any lockout, record last login
		lastLogin := time.Now()
		zeroAttemptSQL := `UPDATE users SET attempts = 0, last_login = ? WHERE username = ?;`
		_, err := a.db.Exec(zeroAttemptSQL, lastLogin, username)
		if err != nil {
			slog.Error("Update failed", "error", err)
			fmt.Println("AUTH FAILURE")
			return
		}
	}

	//Post login commmAND completion
	a.rl.Config.AutoComplete = postLoginCompleter

	// we have to populate the session here
	loggedInAt := time.Now()
	expiresAt := loggedInAt.Add(TIMEOUT)

	func() {
		a.session.UserID = int64(singleUser.ID)
		a.session.Username = username
		a.session.ExpiresAt = expiresAt
		a.session.LastLogin = singleUser.LastLogin
	}()

	// Display the Post login Message ..

	printUserDetails(&singleUser, a.session)

}

func (a *App) handleRegister() {

	a.rl.SetPrompt("Username: ")
	username, _ := a.rl.Readline()

	var lastID int64
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
		slog.Error("HashPassword Failure", err)
		fmt.Println("Registration Failed !!")
	}

	// create the record
	// fields must  userId, username, hashedpassword , failedAttempts , TimeBlocked
	insertSQL := `INSERT INTO users (username, password, created_at) VALUES (?, ?, ?);`

	result, err := a.db.Exec(insertSQL, username, hashedPassword, time.Now())
	if err != nil {
		slog.Error("Insert Failed for the Register Handle (likely duplicate Password or Username): ", err)
		fmt.Println("Registration Failed !!")
		return
	} else {
		lastID, _ = result.LastInsertId()
		fmt.Printf("Registered user with Id %d\n", lastID)

	}
	// populate the session here i would say soo that would be the helpful thing

	a.session = &Session{UserID: lastID, Username: username, LastLogin: nil}
	func() {
		a.session.UserID = lastID
		a.session.Username = username
		a.session.LastLogin = nil
	}()
	a.rl.Config.AutoComplete = postLoginCompleter

}

/* Post Login Methods */

func (a *App) handleWhoami() {

	fmt.Println(a.session.Username)
}

func (a *App) handleLogout() {
	a.session = nil
	a.rl.Config.AutoComplete = preLoginCompleter
}

func (a *App) handleEnable2Fa()

func (a *App) handleDisable2Fa()

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

	if s.LastLogin != nil {
		fmt.Printf("Last login:    %s\n", s.LastLogin.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Last login:    (first login)")
	}
}

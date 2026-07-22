package cli

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/pquerna/otp/totp"
	_ "modernc.org/sqlite"
)

/*Pre Login Methods */

func (a *App) handleLogin() {
	a.rl.SetPrompt("Username: ")
	username, err := a.rl.Readline()
	if err != nil {
		fmt.Println("Auth Failure!!")
		return
	}
	passwordBytes, err := a.rl.ReadPassword("Password")
	if err != nil {
		fmt.Println("Auth Failure!!")
	}
	password := string(passwordBytes)

	userDetailSQL := `SELECT  * FROM users where username = ?;`

	var singleUser User
	err = a.db.Get(&singleUser, userDetailSQL, username)
	if err == sql.ErrNoRows {
		slog.Error("User Not Found")
		fmt.Println("User", username, "Not registered!!")
		return // user doesnt exist so exit the handler
	} else if err != nil {
		slog.Error("Database error", "error", err)
		fmt.Println("AUTH FAILURE !!")
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
		if attempts+1 >= globalTryLimit {
			// block: reset attempts to 0 once blocked, set the lockout window
			attempts = 0
			blockedUntil := time.Now().Add(15 * time.Minute)
			blockSQL := `UPDATE users SET attempts = ?, blocked_time = ? WHERE username = ?;`
			_, err := a.db.Exec(blockSQL, attempts, blockedUntil, username)
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
	// Post password check for the 2 factor authentication
	if singleUser.MFAEnabled {
		a.rl.SetPrompt("2FA code: ")
		code, _ := a.rl.Readline()
		if !totp.Validate(code, singleUser.MFASecret) {
			fmt.Println("Invalid 2FA code")
			return
		}
	}
	//Post login commmAND completion
	a.rl.Config.AutoComplete = postLoginCompleter

	// we have to populate the session here
	loggedInAt := time.Now()
	expiresAt := loggedInAt.Add(SessionTimeout)

	a.session = &Session{
		UserID:    int64(singleUser.ID),
		Username:  singleUser.Username,
		ExpiresAt: expiresAt,
	}

	// Display the Post login Message ..

	printUserDetails(&singleUser, a.session)

}

func (a *App) handleRegister() {

	a.rl.SetPrompt("Username: ")
	username, err := a.rl.Readline()
	if err != nil {
		fmt.Println("Registration Failed !!")
		return
	}

	var lastID int64
	var confirmPassword string
	for {
		passwordBytes, err := a.rl.ReadPassword("Password: ")
		if err != nil {
			fmt.Println("Registration Failed !!")
			return
		}
		password := string(passwordBytes)
		confirmPasswordBytes, _ := a.rl.ReadPassword("Confirm Password: ")
		confirmPassword := string(confirmPasswordBytes)

		if password != confirmPassword {
			fmt.Println("Password Doesnt Matched")
			continue

		} else {
			break
		}
	}
	hashedPassword, err := hashPassword(confirmPassword)
	if err != nil {
		slog.Error("HashPassword Failure", "error", err)
		fmt.Println("Registration Failed !!")
	}

	// create the record
	// fields must  userId, username, hashedpassword , failedAttempts , TimeBlocked
	insertSQL := `INSERT INTO users (username, password, created_at) VALUES (?, ?, ?);`

	result, err := a.db.Exec(insertSQL, username, hashedPassword, time.Now())
	if err != nil {
		slog.Error("Insert Failed for the Register Handle (likely duplicate Password or Username): ", "error", err)
		fmt.Println("Registration Failed !!")
		return
	} else {
		lastID, _ = result.LastInsertId()
		fmt.Printf("Registered user with Id %d\n", lastID)

	}

	// change the completer to the Post Login based
	a.rl.Config.AutoComplete = preLoginCompleter

}

/* Post Login Methods */

func (a *App) handleWhoami() {
	var user User

	err := a.db.Get(
		&user,
		"SELECT * FROM users WHERE id = ?",
		a.session.UserID,
	)

	if err != nil {
		fmt.Println("Unable to fetch user.")
		return
	}
	printUserDetails(&user, a.session)
}

func (a *App) handleLogout() {
	a.session = nil
	a.rl.Config.AutoComplete = preLoginCompleter
	fmt.Println("Logged out")
}

func (a *App) handleEnable2Fa() {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "go-AUTH",
		AccountName: a.session.Username,
	})
	if err != nil {
		slog.Error("totp generate failed", "error", err)
		fmt.Println("Could not enable 2FA")
		return
	}

	fmt.Println("Secret (add this to your authenticator app):", key.Secret())

	a.rl.SetPrompt("Current 2FA code: ")
	code, err := a.rl.Readline()
	if err != nil {
		fmt.Println("Could not enable 2FA")
		return
	}
	if !totp.Validate(code, key.Secret()) {
		fmt.Println("Incorrect code. 2FA not enabled.")
		return
	}

	updateSQL := `UPDATE users SET mfa_enabled = ?, mfa_secret = ? WHERE username = ?`
	if _, err := a.db.Exec(updateSQL, true, key.Secret(), a.session.Username); err != nil {
		slog.Error("failed to save mfa secret", "error", err)
		fmt.Println("Could not enable 2FA")
		return
	}
	// a.session.MFAEnabled = true
	fmt.Println("2FA enabled.")
}

func (a *App) handleDisable2Fa() {

	updateSQL := `
		UPDATE users
		SET mfa_enabled = ?, mfa_secret = ?
		WHERE username = ?;
	`

	_, err := a.db.Exec(updateSQL, false, "", a.session.Username)
	if err != nil {
		slog.Error("failed to disable 2FA", "error", err)
		fmt.Println("Could not disable 2FA")
		return
	}

	fmt.Println("2FA disabled.")
}

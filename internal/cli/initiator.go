package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	// "database/sql"

	"github.com/chzyer/readline"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Session struct {
	UserID     int64
	Username   string
	ExpiresAt  time.Time
	MFAenabled bool
	LastLogin  *time.Time
}

var db *sqlx.DB
var TIMEOUT = 5 * time.Minute

func initDB() {

}

func Initiate() {

	/* DB INITIALIZATION
	sqlx module used for the sql connection, because of its presistent connection check ,
	better than normal 'Open()' and it also provide best wrapper over the 'database/sql'
	to retrieve element over the struct in the variable and easy to use and manage
	*/

	db, err := sqlx.Connect("sqlite", "app.db")
	if err != nil {
		slog.Error("Failed to open database: ", err)
		// some error print and exit
	}
	defer db.Close()

	// 2. Create a table
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT UNIQUE NOT NULL,
		create_at DATETIME
		mfa_enabled BOOL
		last_login DATETIME,
		attempts INTEGER,
		blocked_time DATETIME
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		slog.Error("Failed to create table: %v", err)
		//  Some error and exit
	}
	/*------------------------------------ */

	// CLI Related variable initialization
	var currentSession *Session = &Session{}
	var handlers func()
	var ok bool

	// Readline instance creation for the current tErminal
	rl, err := readline.New("> ")
	if err != nil {
		slog.Error("Readline Error", err)
		fmt.Println("Fail To setup the CLI!!")
		return
	}
	defer rl.Close()

	// App Object Construction
	app := &App{db: db, session: currentSession, rl: rl}

	// make sure app struct populated with necessary things
	preLoginCmds := map[string]func(){
		"register": app.handleRegister,
		"login":    app.handleLogin,
		"help":     handleHelp,
		"exit":     func() { os.Exit(0) }}

	postLoginCmds := map[string]func(){
		"whoami":      app.handleWhoami,
		"enable-2fa":  app.handleEnable2Fa,
		"disable-2fa": app.handleDisable2Fa,
		"logout":      app.handleLogout,
		"help":        handlePostHelp}

	/*-------------MAIN-LOOP-----------*/

	for {
		line, err := rl.Readline()
		if err != nil {
			slog.Error("Readline Error")
			fmt.Println("")
			return
		} // log
		cmd := strings.Fields(strings.TrimSpace(line))
		if len(cmd) > 1 {
			slog.Error("More Than one command")
			fmt.Println("Please Enter only Allowed commands!! use 'help'..")
			return
			//log
		}

		/*  Session Management */
		if currentSession == nil {
			//  do we have to say something or we have to show the comamnds or what
			// or lastLogin thing
			handlers, ok = preLoginCmds[cmd[0]]
		} else {
			// currentSession.ExpiresAt = timeout.Add(currentSession.LoggedInAt.Sub(time.Time{}))

			if time.Now().After(currentSession.ExpiresAt) {
				currentSession = nil
				continue
			}

			handlers, ok = postLoginCmds[cmd[0]]

		}

		if !ok {
			fmt.Printf("Unknown command!! '%v'", cmd[0])
			continue
		}

		err = handlers()
		if err != nil {
		}

		rl.SetPrompt("> ") // Precautionary doing so , for any handler that made changes
	}

}

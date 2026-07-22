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
	UserID    int64
	Username  string
	ExpiresAt time.Time
	// MFAenabled bool
	// LastLogin  *time.Time
}

var preLoginCompleter = readline.NewPrefixCompleter(
	readline.PcItem("register"),
	readline.PcItem("login"),
	readline.PcItem("help"),
	readline.PcItem("exit"),
)

var postLoginCompleter = readline.NewPrefixCompleter(
	readline.PcItem("whoami"),
	readline.PcItem("enable-2fa"),
	readline.PcItem("disable-2fa"),
	readline.PcItem("logout"),
	readline.PcItem("help"),
)
var SessionTimeout = 5 * time.Minute

func Initiate() {
	var db *sqlx.DB

	/* DB INITIALIZATION
	sqlx module used for the sql connection, because of its presistent connection check ,
	better than normal 'Open()' and it also provide best wrapper over the 'database/sql'
	to retrieve element over the struct in the variable and easy to use and manage
	*/

	db, err := sqlx.Connect("sqlite", "/app/data/app.db")
	if err != nil {
		slog.Error("Failed to open database ", "error: ", err)
		// some error print and exit
	}
	defer db.Close()

	// 2. Create a table
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT  NOT NULL,
		created_at DATETIME,
		mfa_enabled NOT NULL DEFAULT 0,
		mfa_secret TEXT NOT NULL DEFAULT '' ,
		last_login DATETIME,
		attempts INTEGER DEFAULT 0,
		blocked_time DATETIME
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		slog.Error("Failed to create table ", "error :", err)
		//  Some error and exit
		fmt.Println("CLI Init-Setup Failed")
		os.Exit(1)
	}

	// CLI Related variable initialization
	var handlers func()

	// Readline instance creation for the current tErminal
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       ">> ",
		AutoComplete: preLoginCompleter,
	})
	if err != nil {
		slog.Error("Readline Error", "error: ", err)
		fmt.Println("Fail To setup the CLI!!")
		return
	}
	defer rl.Close()

	// App Object Construction
	app := &App{db: db, rl: rl}

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

	// call the banner
	banner()

	/*-------------MAIN-LOOP-----------*/

	for {

		// session Expiry  Check
		if app.session != nil &&
			time.Now().After(app.session.ExpiresAt) {

			fmt.Println("Session expired.")
			app.session = nil
		}

		line, err := rl.Readline()
		if err != nil {
			slog.Error("Readline Error")
			fmt.Println("")
			return
		}
		cmd := strings.Fields(strings.TrimSpace(line))

		if len(cmd) == 0 {
			continue // user just hit Enter on an empty line — nothing to do, loop again
		}

		if len(cmd) > 1 {
			slog.Error("More Than one command")
			fmt.Println("Please Enter only Allowed commands!! use 'help'..")
			continue
		}

		/*  Session Management */
		if app.session == nil {
			handlers = preLoginCmds[cmd[0]]
			app.rl.Config.AutoComplete = preLoginCompleter
		} else {
			app.rl.Config.AutoComplete = postLoginCompleter
			handlers = postLoginCmds[cmd[0]]

		}

		if handlers == nil {
			fmt.Printf("Unknown command!! '%v'\n", cmd[0])
			continue
		}

		handlers()

		rl.SetPrompt(">> ") // Precautionary doing so , for any handler that made changes
		fmt.Println()
	}

}

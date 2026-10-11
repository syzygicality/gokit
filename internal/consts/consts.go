package consts

const DatabasePath = "~/.gokit/aliases"

const APIPrefix = "https://pkg.go.dev/v1/"

const VerboseMode = false

const DatabaseDirPerm = 0o755

const SearchDefaultLimit = 26

// AliasMaxLength matches the alias column width in schema.sql.
const AliasMaxLength = 32

const (
	HelpCommand   = "help"
	ListCommand   = "list"
	SaveCommand   = "save"
	RemoveCommand = "remove"
	SearchCommand = "search"
)

// Commands lists every subcommand name; aliases cannot use these since the command would shadow them.
var Commands = []string{HelpCommand, ListCommand, SaveCommand, RemoveCommand, SearchCommand}

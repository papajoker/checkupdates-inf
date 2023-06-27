package theme

var (
	ColorNone    = "\033[0m"
	ColorBold    = "\033[1m"
	ColorBlue    = "\033[0;94m"
	ColorGreen   = "\033[0;92m"
	ColorGray    = "\033[0;90m"
	ColorRed     = "\033[0;91m"
	ColorWarning = "\033[93m"
)

func Reset() {
	ColorNone = ""
	ColorBold = ""
	ColorBlue = ""
	ColorGreen = ""
	ColorGray = ""
	ColorRed = ""
	ColorWarning = ""
}

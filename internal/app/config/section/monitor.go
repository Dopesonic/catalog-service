package section

type Monitor struct {
	LogLevel    string `split_words:"true" default:"debug"`
	Environment string `split_words:"true" default:"development"`
}

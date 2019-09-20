package promptx

import (
	"text/template"

	"github.com/mritd/promptx/utils"
	"github.com/mritd/readline"
)

const (
	DefaultPrompt         = "»"
	DefaultErrorMsgPrefix = "✘ "
	DefaultAskTpl         = "{{ . | cyan }} "
	DefaultPromptTpl      = "{{ . | green }} "
	DefaultInvalidTpl     = "{{ . | red }} "
	DefaultValidTpl       = "{{ . | green }} "
	DefaultErrorMsgTpl    = "{{ . | red }} "
)

type Prompt struct {
	Config
	Ask     string
	Prompt  string
	FuncMap template.FuncMap

	isFirstRun bool

	ask      *template.Template
	prompt   *template.Template
	valid    *template.Template
	invalid  *template.Template
	errorMsg *template.Template
}

type Config struct {
	AskTpl        string
	PromptTpl     string
	ValidTpl      string
	InvalidTpl    string
	ErrorMsgTpl   string
	CheckListener func(line []rune) error
}

func NewDefaultConfig(check func(line []rune) error) Config {
	return Config{
		AskTpl:        DefaultAskTpl,
		PromptTpl:     DefaultPromptTpl,
		InvalidTpl:    DefaultInvalidTpl,
		ValidTpl:      DefaultValidTpl,
		ErrorMsgTpl:   DefaultErrorMsgTpl,
		CheckListener: check,
	}
}

func NewDefaultPrompt(check func(line []rune) error, ask string) Prompt {
	return Prompt{
		Ask:     ask,
		Prompt:  DefaultPrompt,
		FuncMap: FuncMap,
		Config:  NewDefaultConfig(check),
	}
}

func (p *Prompt) prepareTemplates() {

	var err error
	p.ask, err = template.New("").Funcs(FuncMap).Parse(p.AskTpl)
	utils.CheckAndExit(err)
	p.prompt, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl)
	utils.CheckAndExit(err)
	p.valid, err = template.New("").Funcs(FuncMap).Parse(p.ValidTpl)
	utils.CheckAndExit(err)
	p.invalid, err = template.New("").Funcs(FuncMap).Parse(p.InvalidTpl)
	utils.CheckAndExit(err)
	p.errorMsg, err = template.New("").Funcs(FuncMap).Parse(p.ErrorMsgTpl)
	utils.CheckAndExit(err)

}

func (p *Prompt) Run() string {
	p.isFirstRun = true
	p.prepareTemplates()

	displayPrompt := append(utils.Render(p.prompt, p.Prompt), utils.Render(p.ask, p.Ask)...)
	validPrompt := append(utils.Render(p.valid, p.Prompt), utils.Render(p.ask, p.Ask)...)
	invalidPrompt := append(utils.Render(p.invalid, p.Prompt), utils.Render(p.ask, p.Ask)...)

	l, err := readline.NewEx(&readline.Config{
		Prompt:                 string(displayPrompt),
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
	})
	utils.CheckAndExit(err)

	filterInput := func(r rune) (rune, bool) {

		switch r {
		// block CtrlZ feature
		case readline.CharCtrlZ:
			return r, false
		default:
			return r, true
		}
	}

	l.Config.FuncFilterInputRune = filterInput

	l.Config.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		// Real-time verification
		if err = p.CheckListener(line); err != nil {
			l.SetPrompt(string(invalidPrompt))
			l.Refresh()
		} else {
			l.SetPrompt(string(validPrompt))
			l.Refresh()
		}
		return nil, 0, false
	})
	defer func() { _ = l.Close() }()

	// read line
	for {
		if !p.isFirstRun {
			_, err := l.Write([]byte(moveUp))
			utils.CheckAndExit(err)
		}
		s, err := l.Readline()
		utils.CheckAndExit(err)
		if err = p.CheckListener([]rune(s)); err != nil {
			l.Write([]byte(clearLine))
			l.Write([]byte(string(utils.Render(p.errorMsg, DefaultErrorMsgPrefix+err.Error()))))
			p.isFirstRun = false
		} else {
			return s
		}
	}
}

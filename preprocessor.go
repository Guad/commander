package commander

import "strings"

// Preprocessor parses and changes the command args before the main parser takes them. Return false on Process if you have to cancel the command.
type Preprocessor interface {
	Process([]string) ([]string, bool)
}

// TelegramPreprocessor removes the part after the @ only if it matches the bot's name.
type TelegramPreprocessor struct {
	BotName string
}

func (t *TelegramPreprocessor) Process(args []string) ([]string, bool) {
	prefix := args[0]

	if strings.ContainsRune(prefix, '@') {
		i := strings.IndexRune(prefix, '@')
		runes := []rune(prefix)
		name := string(runes[i+1:])
		prefix = string(runes[:i])

		if name == t.BotName {
			args[0] = prefix
		} else {
			return nil, false
		}
	}

	return args, true
}

// Ensure TelegramPreprocessor adheres to the interface.
var _ Preprocessor = &TelegramPreprocessor{}

// IRCPreprocessor makes commands starting with ! work
type IRCPreprocessor struct {
}

func (t *IRCPreprocessor) Process(args []string) ([]string, bool) {
	prefix := args[0]

	if len(prefix) > 0 && prefix[0] == '!' {
		args[0] = string(append([]rune{'/'}, []rune(prefix[1:])...))
		return args, true
	}

	return nil, false
}

// Ensure TelegramPreprocessor adheres to the interface.
var _ Preprocessor = &IRCPreprocessor{}

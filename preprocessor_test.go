package commander

import "testing"

func TestTelegramPreprocessor(t *testing.T) {
	cmd := New()

	cmd.Preprocessor = &TelegramPreprocessor{
		BotName: "Commander",
	}

	hand := func(ctx Context) error {
		return nil
	}

	cmd.Command("/test", hand)

	succ, _ := cmd.Execute("/test")

	if !succ {
		t.Fail()
	}

	succ, _ = cmd.Execute("/test@Commander")

	if !succ {
		t.Fail()
	}

	succ, _ = cmd.Execute("/test@OtherBot")

	if succ {
		t.Fail()
	}
}

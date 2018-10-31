package commander

import "testing"

func simpleCmdHandler(ctx Context) error {
	return nil
}

func TestSimpleCommand(t *testing.T) {
	c := New()

	c.Command("/test", simpleCmdHandler)

	succ, err := c.Execute("/test")

	if !succ || err != nil {
		t.Fail()
	}

	succ, err = c.Execute("/tes")

	if succ {
		t.Fail()
	}

	succ, err = c.Execute("/test arg1")

	if !succ || err != ErrTooManyArgs {
		t.Fail()
	}
}

func TestSimpleArgs(t *testing.T) {
	c := New()

	c.Command("/test {arg1}", simpleCmdHandler)

	succ, err := c.Execute("/test hello")

	if !succ || err != nil {
		t.Fail()
	}

	succ, err = c.Execute("/test")

	if !succ || err != ErrNotEnoughArgs {
		t.Fail()
	}

	succ, err = c.Execute("/test hello world")

	if !succ || err != ErrTooManyArgs {
		t.Fail()
	}
}

func TestTypedArgs(t *testing.T) {
	c := New()

	c.Command("/test {arg1:int}", simpleCmdHandler)

	succ, err := c.Execute("/test 10")

	if !succ || err != nil {
		t.Log("Should succeed, got succ:", succ, "err:", err)
		t.Fail()
	}

	succ, err = c.Execute("/test hello")

	if !succ || err != ErrArgSyntaxError {
		t.Log("[1] Should fail with ErrArgSyntaxError, got succ:", succ, "err:", err)
		t.Fail()
	}
}

func TestPlusArg(t *testing.T) {
	c := New()

	c.Command("/test {arg1} {arg2+}", simpleCmdHandler)

	succ, err := c.Execute("/test")

	if !succ || err != ErrNotEnoughArgs {
		t.Fail()
	}

	succ, err = c.Execute("/test hello")

	if !succ || err != ErrNotEnoughArgs {
		t.Fail()
	}

	succ, err = c.Execute("/test hello world")

	if !succ || err != nil {
		t.Fail()
	}

	succ, err = c.Execute("/test hello world foo bar")

	if !succ || err != nil {
		t.Fail()
	}
}

func TestStarArg(t *testing.T) {
	c := New()

	c.Command("/test {arg1} {arg2*}", simpleCmdHandler)

	succ, err := c.Execute("/test")

	if !succ || err != ErrNotEnoughArgs {
		t.Fail()
	}

	succ, err = c.Execute("/test hello")

	if !succ || err != nil {
		t.Fail()
	}

	succ, err = c.Execute("/test hello world")

	if !succ || err != nil {
		t.Fail()
	}

	succ, err = c.Execute("/test hello world foo bar")

	if !succ || err != nil {
		t.Fail()
	}
}

func TestOptionalArg(t *testing.T) {
	c := New()

	c.Command("/test {arg1} {arg2?}", simpleCmdHandler)

	succ, err := c.Execute("/test")

	if !succ || err != ErrNotEnoughArgs {
		t.Fail()
	}

	succ, err = c.Execute("/test hello")

	if !succ || err != nil {
		t.Fail()
	}

	succ, err = c.Execute("/test hello world")

	if !succ || err != nil {
		t.Fail()
	}

	succ, err = c.Execute("/test hello world foo bar")

	if !succ || err != ErrTooManyArgs {
		t.Fail()
	}
}

func TestMiddleware(t *testing.T) {
	c := New()

	c.Use(func(next Handler) Handler {
		return func(ctx Context) error {
			ctx.AddArg("FIRST", 10)
			return next(ctx)
		}
	})

	c.Use(func(next Handler) Handler {
		return func(ctx Context) error {
			if ctx.ArgInt("FIRST") != 10 {
				t.Fail()
			}
			return next(ctx)
		}
	})

	c.Command("/test", simpleCmdHandler)

	succ, err := c.Execute("/test")

	if !succ || err != nil {
		t.Fail()
	}
}

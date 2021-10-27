package inbox

type offset struct {
}

func NewOffseter() Offseter {
	return &offset{}
}

// Status message dispatched.
func (off *offset) Status() bool {
	return true
}

// Confirm confirm message.
func (off *offset) Confirm() {}

func (off *offset) Commit() error {
	return nil
}

func (off *offset) AutoCommit() bool {
	return true
}

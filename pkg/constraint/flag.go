package constraint

const (
	EnabledFlagSelf       uint8 = 1
	EnabledFlagSearch     uint8 = 2
	EnabledFlagTimeSeries uint8 = 4
)

type enableFlag struct {
	value uint8
}

func NewEnableFlag() *enableFlag {
	return &enableFlag{}
}

func (ef *enableFlag) EnableAll() {
	ef.value = ef.value | EnabledFlagSelf | EnabledFlagSearch | EnabledFlagTimeSeries
}

func (ef *enableFlag) Enabled() bool {
	return EnabledFlagSelf == (ef.value & EnabledFlagSelf)
}

func (ef *enableFlag) Enable(flag bool) bool {
	return ef.enable(flag, EnabledFlagSelf)
}

func (ef *enableFlag) Searchable() bool {
	return EnabledFlagSearch == (ef.value & EnabledFlagSearch)
}

func (ef *enableFlag) EnableSearch(flag bool) bool {
	return ef.enable(flag, EnabledFlagSearch)
}

func (ef *enableFlag) TSEnabled() bool {
	return EnabledFlagTimeSeries == (ef.value & EnabledFlagTimeSeries)
}

func (ef *enableFlag) EnableTS(flag bool) bool {
	return ef.enable(flag, EnabledFlagTimeSeries)
}

func (ef *enableFlag) enable(flag bool, flagValue uint8) bool {
	retFlag := ef.value
	if flag {
		ef.value = ef.value | flagValue
	} else {
		ef.value = ef.value & (0xff ^ flagValue)
	}

	return flagValue == (retFlag & flagValue)
}

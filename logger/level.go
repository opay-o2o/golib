package logger

type Level byte

const (
	DisableLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

var Levels = map[Level]*LevelMeta{
	DisableLevel: {
		Name:      "disable",
		Alias:     []string{"disabled"},
		RawText:   "",
		ColorText: "",
	},
	FatalLevel: {
		Name:      "fatal",
		RawText:   "[FTAL]",
		ColorText: RedBackground("[FTAL]"),
	},
	ErrorLevel: {
		Name:      "error",
		RawText:   "[ERRO]",
		ColorText: Red("[ERRO]"),
	},
	WarnLevel: {
		Name:      "warn",
		Alias:     []string{"warning"},
		RawText:   "[WARN]",
		ColorText: Purple("[WARN]"),
	},
	InfoLevel: {
		Name:      "info",
		RawText:   "[INFO]",
		ColorText: LightGreen("[INFO]"),
	},
	DebugLevel: {
		Name:      "debug",
		RawText:   "[DBUG]",
		ColorText: Yellow("[DBUG]"),
	},
}

type LevelMeta struct {
	Name      string
	Alias     []string
	RawText   string
	ColorText string
}

func (m *LevelMeta) Text(enableColor bool) string {
	if enableColor {
		return m.ColorText
	}
	return m.RawText
}

func GetLevel(levelName string) Level {
	for level, meta := range Levels {
		if meta.Name == levelName {
			return level
		}

		for _, altName := range meta.Alias {
			if altName == levelName {
				return level
			}
		}
	}

	return DisableLevel
}

func GetLevelText(level Level, enableColor bool) string {
	if meta, ok := Levels[level]; ok {
		return meta.Text(enableColor)
	}

	return ""
}

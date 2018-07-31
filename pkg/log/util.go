package log

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// LazyEval can be used to evaluate an argument under a correct log level.
type LazyEval func() string

func (l LazyEval) String() string {
	return l()
}

// DoLazyEval returns LazyEval. Unnecessary evalution can be prevented by using
// "%v" format string,
func DoLazyEval(c func() string) LazyEval {
	return LazyEval(c)
}

func ParseLevels(compositeLevel string) (*Levels, error) {
	if !strings.Contains(compositeLevel, ",") && !strings.Contains(compositeLevel, "=") {
		level := strings.TrimSpace(compositeLevel)
		if _, err := logrus.ParseLevel(level); err != nil {
			return nil, err
		}
		return &Levels{
			Default: level,
			Module:  map[string]string{},
		}, nil
	}

	var defaultLevel string
	moduleLevelMap := make(map[string]string, 0)

	for _, moduleLevelPair := range strings.Split(compositeLevel, ",") {
		if false == strings.Contains(moduleLevelPair, "=") {
			level := strings.TrimSpace(moduleLevelPair)
			if _, err := logrus.ParseLevel(level); err != nil {
				return nil, err
			}
			defaultLevel = level
		} else {
			fields := strings.Split(moduleLevelPair, "=")
			if len(fields) != 2 {
				return nil, fmt.Errorf("invalid log format for module; %s, must contain only one =", moduleLevelPair)
			}
			module, level := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
			if _, err := logrus.ParseLevel(level); err != nil {
				return nil, err
			}
			moduleLevelMap[strings.ToLower(module)] = level
		}
	}

	return &Levels{
		Default: defaultLevel,
		Module:  moduleLevelMap,
	}, nil
}

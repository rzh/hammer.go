package scenario

import (
	"errors"
	"math/rand"
)

type Profile interface {
	InitFromFile(string)
	InitFromCode()
	NextCall(*rand.Rand) (*Call, error)
	CustomizedReport() string
}

var scenarios = make(map[string]func(int) (Profile, error))

func Register(name string, scenario func(int) (Profile, error)) {
	scenarios[name] = scenario
}

func New(scenarioName string, sessionSize int) (Profile, error) {
	if scenario, ok := scenarios[scenarioName]; ok {
		return scenario(sessionSize)
	}

	return nil, errors.New("scenario is not registered")
}

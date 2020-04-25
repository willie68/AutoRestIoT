package worker

import (
	"errors"

	"github.com/qntfy/kazaam"
)

var ErrRuleNotDefined = errors.New("Rule not defined")

var rules map[string]kazaam.Kazaam

func init() {
	rules = make(map[string]kazaam.Kazaam)
}

func registerTransformRule(name string, config string) error {
	k, err := kazaam.NewKazaam(config)
	if err != nil {
		log.Alertf("Unable to transform message %v", err)
		return err
	}
	rules[name] = *k
	return nil
}

func transformJSON(name, json string) (string, error) {
	k, ok := rules[name]
	if !ok {
		return "", ErrRuleNotDefined
	}
	out, transformError := k.TransformJSONStringToString(json)
	if transformError != nil {
		log.Alertf("Unable to transform message %v", transformError)
		return "", transformError
	}

	return out, nil
}

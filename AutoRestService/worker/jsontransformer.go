package worker

import (
	"errors"

	"github.com/qntfy/kazaam"
)

var ErrRuleNotDefined = errors.New("Rule not defined")

var rules map[string]kazaam.Kazaam
var kazaamConfig kazaam.Config

func init() {
	rules = make(map[string]kazaam.Kazaam)
	kazaamConfig = kazaam.NewDefaultConfig()
}

func registerTransformRule(name string, config string) error {
	k, err := kazaam.New(config, kazaamConfig)
	if err != nil {
		log.Alertf("Unable to transform message %v", err)
		return err
	}
	rules[name] = *k
	return nil
}

func TransformJSON(name string, json []byte) ([]byte, error) {
	k, ok := rules[name]
	if !ok {
		return []byte{}, ErrRuleNotDefined
	}
	out, transformError := k.TransformInPlace(json)
	if transformError != nil {
		log.Alertf("Unable to transform message %v", transformError)
		return []byte{}, transformError
	}

	return out, nil
}

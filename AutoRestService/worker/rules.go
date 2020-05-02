package worker

import (
	"errors"
	"fmt"
	"sync"

	"github.com/qntfy/kazaam"
)

var ErrRuleNotDefined = errors.New("Rule not defined")

type RuleList struct {
	rules     map[string]kazaam.Kazaam
	rulesSync sync.Mutex
}

var Rules = RuleList{
	rules: make(map[string]kazaam.Kazaam),
}

var kazaamConfig kazaam.Config

func init() {
	kazaamConfig = kazaam.NewDefaultConfig()
}

//GetRuleName building the rule namespace name
func GetRuleNsName(backendName string, rulename string) string {
	return fmt.Sprintf("%s.%s", backendName, rulename)
}

func (r *RuleList) GetRulelist() []string {
	list := make([]string, 0)
	r.rulesSync.Lock()
	for k, _ := range r.rules {
		list = append(list, k)
	}
	r.rulesSync.Unlock()
	return list
}

func (r *RuleList) Register(backendName, rulename string, config string) error {
	name := GetRuleNsName(backendName, rulename)
	k, err := kazaam.New(config, kazaamConfig)
	if err != nil {
		log.Alertf("Unable to transform message %v", err)
		return err
	}
	r.rulesSync.Lock()
	r.rules[name] = *k
	r.rulesSync.Unlock()
	return nil
}

func (r *RuleList) Deregister(backendName, rulename string) error {
	name := GetRuleNsName(backendName, rulename)
	r.rulesSync.Lock()
	delete(r.rules, name)
	r.rulesSync.Unlock()
	return nil
}

func (r *RuleList) TransformJSON(backendName, rulename string, json []byte) ([]byte, error) {
	name := GetRuleNsName(backendName, rulename)
	r.rulesSync.Lock()
	k, ok := r.rules[name]
	r.rulesSync.Unlock()
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

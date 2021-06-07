// Code generated by "genfeature -receiver h Handler"; DO NOT EDIT.

package xtime

import (
	"mellium.im/xmpp/disco/info"
)

// A list of service discovery features that are supported by this package.
var (
	Feature = info.Feature{Var: NS}
)

// ForFeatures implements info.FeatureIter.
func (h Handler) ForFeatures(node string, f func(info.Feature) error) error {
	if node != "" {
		return nil
	}
	var err error
	err = f(Feature)
	if err != nil {
		return err
	}
	return nil
}

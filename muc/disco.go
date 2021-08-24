// Code generated by "genfeature -receiver *Client"; DO NOT EDIT.

package muc

import (
	"mellium.im/xmpp/disco/info"
)

// A list of service discovery features that are supported by this package.
var (
	Feature = info.Feature{Var: NS}
)

// ForFeatures implements info.FeatureIter.
func (*Client) ForFeatures(node string, f func(info.Feature) error) error {
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
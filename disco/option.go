// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package disco

// An Option is used to configure new registries.
type Option func(*Registry)

// Identity adds an identity to the registry.
//
// Identities are described by XEP-0030:
//
//     An entity's identity is broken down into its category (server, client,
//     gateway, directory, etc.) and its particular type within that category
//     (IM server, phone vs. handheld client, MSN gateway vs. AIM gateway, user
//     directory vs. chatroom directory, etc.). This information helps
//     requesting entities to determine the group or "bucket" of services into
//     which the entity is most appropriately placed (e.g., perhaps the entity
//     is shown in a GUI with an appropriate icon).
func Identity(category, typ, name, lang string) Option {
	return func(r *Registry) {
		if r.identities == nil {
			r.identities = make(map[identity]string)
		}
		r.identities[identity{
			Category: category,
			Type:     typ,
			XMLLang:  lang,
		}] = name
	}
}

// Feature adds a feature to the registry.
//
// Features are described by XEP-0030:
//
//     This information helps requesting entities determine what actions are
//     possible with regard to this entity (registration, search, join, etc.),
//     what protocols the entity supports, and specific feature types of
//     interest, if any (e.g., for the purpose of feature negotiation).
func Feature(name string) Option {
	return func(r *Registry) {
		r.features[name] = struct{}{}
	}
}

// Merge adds all features from r into the registry the option is applied to.
func Merge(r *Registry) Option {
	return func(dst *Registry) {
		for k := range r.features {
			dst.features[k] = struct{}{}
		}
		for k, v := range r.identities {
			dst.identities[k] = v
		}
	}
}

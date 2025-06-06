// Copyright 2023 Juan Pablo Tosso and the OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

package collections

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ad3n/seclang/internal/corazarules"
	"github.com/corazawaf/coraza/v3/collection"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/corazawaf/coraza/v3/types/variables"
)

// NamedCollection is a Collection that also keeps track of names.
type NamedCollection struct {
	*Map
}

var _ collection.Map = &NamedCollection{}

func NewCaseSensitiveNamedCollection(rv variables.RuleVariable) *NamedCollection {
	return &NamedCollection{
		Map: NewCaseSensitiveKeyMap(rv),
	}
}

func NewNamedCollection(rv variables.RuleVariable) *NamedCollection {
	return &NamedCollection{
		Map: NewMap(rv),
	}
}

// Add a value to some key
func (c *NamedCollection) Add(key string, value string) {
	c.Map.Add(key, value)
}

// Set will replace the key's value with this slice
func (c *NamedCollection) Set(key string, values []string) {
	c.Map.Set(key, values)
}

// SetIndex will place the value under the index
// If the index is higher than the current size of the CollectionMap
// it will be appended
func (c *NamedCollection) SetIndex(key string, index int, value string) {
	c.Map.SetIndex(key, index, value)
}

// Remove deletes the key from the CollectionMap
func (c *NamedCollection) Remove(key string) {
	c.Map.Remove(key)
}

func (c *NamedCollection) Len() int {
	return len(c.Map.data)
}

// Data is an internal method used for serializing to JSON
func (c *NamedCollection) Data() map[string][]string {
	result := make(map[string][]string, len(c.data))
	for k, v := range c.data {
		result[k] = make([]string, len(v))
		for i, a := range v {
			result[k][i] = a.value
		}
	}
	return result
}

// Name returns the name for the current CollectionMap
func (c *NamedCollection) Name() string {
	return c.Map.Name()
}

func (c *NamedCollection) Reset() {
	c.Map.Reset()
}

func (c *NamedCollection) Names(rv variables.RuleVariable) collection.Keyed {
	return &NamedCollectionNames{
		variable:   rv,
		collection: c,
	}
}

func (c *NamedCollection) Format(res *strings.Builder) {
	c.Map.Format(res)
}

func (c *NamedCollection) String() string {
	return fmt.Sprint(c.Map)
}

type NamedCollectionNames struct {
	variable   variables.RuleVariable
	collection *NamedCollection
}

func (c *NamedCollectionNames) FindRegex(key *regexp.Regexp) []types.MatchData {
	var res []types.MatchData

	for k, data := range c.collection.Map.data {
		if !key.MatchString(k) {
			continue
		}
		for _, d := range data {
			res = append(res, &corazarules.MatchData{
				Variable_: c.variable,
				Key_:      d.key,
				Value_:    d.key,
			})
		}
	}
	return res
}

func (c *NamedCollectionNames) FindString(key string) []types.MatchData {
	var res []types.MatchData

	for k, data := range c.collection.Map.data {
		if k != key {
			continue
		}
		for _, d := range data {
			res = append(res, &corazarules.MatchData{
				Variable_: c.variable,
				Key_:      d.key,
				Value_:    d.key,
			})
		}
	}
	return res
}

func (c *NamedCollectionNames) Get(key string) []string {
	return c.collection.Map.Get(key)
}

func (c *NamedCollectionNames) FindAll() []types.MatchData {
	var res []types.MatchData
	// Iterates over all the data in the map and adds the key element also to the Key field (The key value may be the value
	//  that is matched, but it is still also the key of the pair and it is needed to print the matched var name)
	for _, data := range c.collection.Map.data {
		for _, d := range data {
			res = append(res, &corazarules.MatchData{
				Variable_: c.variable,
				Key_:      d.key,
				Value_:    d.key,
			})
		}
	}
	return res
}

func (c *NamedCollectionNames) Name() string {
	return c.variable.Name()
}

func (c *NamedCollectionNames) String() string {
	res := strings.Builder{}
	res.WriteString(c.variable.Name())
	res.WriteString(": ")
	firstOccurrence := true
	for _, data := range c.collection.Map.data {
		for _, d := range data {
			if !firstOccurrence {
				res.WriteString(",")
			}
			firstOccurrence = false
			res.WriteString(d.key)
		}
	}
	return res.String()
}

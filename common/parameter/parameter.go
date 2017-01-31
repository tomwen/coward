//  Crypto-Obscured Forwarder
//
//  Copyright (C) 2017 NI Rui <nickriose@gmail.com>
//
//  This file is part of Crypto-Obscured Forwarder.
//
//  Crypto-Obscured Forwarder is free software: you can redistribute it
//  and/or modify it under the terms of the GNU General Public License
//  as published by the Free Software Foundation, either version 3 of
//  the License, or (at your option) any later version.
//
//  Crypto-Obscured Forwarder is distributed in the hope that it will be
//  useful, but WITHOUT ANY WARRANTY; without even the implied warranty
//  of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with Crypto-Obscured Forwarder. If not, see
//  <http://www.gnu.org/licenses/>.

package parameter

// Parameter is the paramster string parser
type Parameter interface {
	Value() *Value
}

// parameter implement Parameter
type parameter struct {
	root      *token
	input     []byte
	nodeSpace int
	maxLevel  int
}

// New creates a new Parameter
func New(input []byte, maxLevel int) (Parameter, error) {
	p := &parameter{
		root:      nil,
		input:     input,
		nodeSpace: 0,
		maxLevel:  maxLevel,
	}

	analyzeErr := p.analyze()

	if analyzeErr != nil {
		return nil, analyzeErr
	}

	return p, nil
}

// isWhiteSpace check if inputted charactor is a white space
func (p *parameter) isWhiteSpace(char byte) bool {
	switch char {
	case ' ':
		fallthrough
	case '\r':
		fallthrough
	case '\n':
		fallthrough
	case '\t':
		return true
	}

	return false
}

// analyze analyzes inputted parameters
func (p *parameter) analyze() error {
	nextStart := 0
	nodeSpace := 1
	lastChar := byte(0)
	root := &token{
		start:    0,
		end:      len(p.input),
		label:    nil,
		fragment: valueFragment,
		symbol:   SymbolValue,
		sub:      tokens{},
	}
	current := &tree{
		token:     root,
		lastLabel: nil,
		fragment:  valueFragment,
		parent:    nil,
		lastStart: 0,
		level:     0,
	}
	treeItem := &tree{
		token: &token{
			start:    nextStart,
			end:      treeInitEnd,
			label:    nil,
			fragment: valueFragment,
			symbol:   SymbolValue,
			sub:      tokens{},
		},
		lastLabel: nil,
		fragment:  valueFragment,
		parent:    nil,
		lastStart: 0,
		level:     current.level,
	}

	for index, charactor := range p.input {
		if treeItem.fragment.needEscape(p.input, index) {
			lastChar = charactor

			continue
		} else if treeItem.isTail(charactor) {
			treeItem.end = index

			current.append(treeItem.token)
			nodeSpace++

			treeItem = &tree{
				token: &token{
					start:    index + 1,
					end:      treeInitEnd,
					label:    current.lastLabel,
					fragment: valueFragment,
					symbol:   SymbolValue,
					sub:      tokens{},
				},
				lastLabel: current.lastLabel,
				fragment:  valueFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level,
			}

			current.lastStart = treeItem.start

			nextStart = index
			lastChar = charactor

			continue
		} else if treeItem.fragment.keepSeeking {
			lastChar = charactor

			continue
		} else if treeItem.isFall(charactor) {
			treeItem.end = index

			current.append(treeItem.token)
			nodeSpace++

			current = treeItem.parent.parent

			treeItem.parent.end = index

			treeItem = &tree{
				token: &token{
					start:    index + 1,
					end:      treeInitEnd,
					label:    current.lastLabel,
					fragment: valueFragment,
					symbol:   SymbolValue,
					sub:      tokens{},
				},
				lastLabel: current.lastLabel,
				fragment:  valueFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level,
			}

			nextStart = index
			lastChar = charactor

			continue
		}

		switch charactor {
		case labelFragment.head:
			// Anti -param-bla check
			if index > 0 && !p.isWhiteSpace(lastChar) {
				nextStart = index
				lastChar = charactor

				continue
			}

			if current.lastStart == treeItem.start &&
				treeItem.fragment.head == labelFragment.head {
				nextStart = index
				lastChar = charactor

				continue
			}

			treeItem.end = index

			current.append(treeItem.token)
			nodeSpace++

			treeItem = &tree{
				token: &token{
					start:    index + 1,
					end:      treeInitEnd,
					label:    current.label,
					fragment: labelFragment,
					symbol:   SymbolLabel,
					sub:      tokens{},
				},
				lastLabel: current.lastLabel,
				fragment:  labelFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level,
			}

			current.lastStart = treeItem.start
			current.lastLabel = treeItem.token
			nextStart = index

		case blockFragment.head:
			if current.lastStart == treeItem.start &&
				treeItem.fragment.head == blockFragment.head {
				nextStart = index
				lastChar = charactor

				continue
			}

			if current.level >= p.maxLevel {
				return newSyntaxErr(
					ErrTreeTokenTooNestDeep, index, p.input)
			}

			treeItem.end = index

			current.append(treeItem.token)
			nodeSpace++

			treeItem = &tree{
				token: &token{
					start:    index + 1,
					end:      treeInitEnd,
					label:    current.lastLabel,
					fragment: blockFragment,
					symbol:   SymbolBlock,
					sub:      tokens{},
				},
				lastLabel: nil,
				fragment:  blockFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level + 1,
			}

			current.lastStart = treeItem.start

			current.append(treeItem.token)
			nodeSpace++

			current = treeItem

			treeItem = &tree{
				token: &token{
					start:    index + 1,
					end:      treeInitEnd,
					label:    nil,
					fragment: valueFragment,
					symbol:   SymbolValue,
					sub:      tokens{},
				},
				lastLabel: nil,
				fragment:  valueFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level,
			}

			current.lastStart = treeItem.start
			nextStart = index

		case doubleQuoteFragment.head:
			// Anti -param"data" and value" check
			if !p.isWhiteSpace(lastChar) {
				nextStart = index
				lastChar = charactor

				continue
			}

			if current.lastStart == treeItem.start &&
				treeItem.fragment.head == doubleQuoteFragment.head {
				nextStart = index
				lastChar = charactor

				continue
			}

			treeItem.end = index

			current.append(treeItem.token)
			nodeSpace++

			treeItem = &tree{
				token: &token{
					start:    index + 1,
					end:      treeInitEnd,
					label:    current.lastLabel,
					fragment: doubleQuoteFragment,
					symbol:   SymbolValue,
					sub:      tokens{},
				},
				lastLabel: current.lastLabel,
				fragment:  doubleQuoteFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level,
			}

			current.lastStart = treeItem.start
			nextStart = index

		case quoteFragment.head:
			// Anti -param"data" and value" check
			if !p.isWhiteSpace(lastChar) {
				nextStart = index
				lastChar = charactor

				continue
			}

			if current.lastStart == treeItem.start &&
				treeItem.fragment.head == quoteFragment.head {
				nextStart = index
				lastChar = charactor

				continue
			}

			treeItem.end = index

			current.append(treeItem.token)
			nodeSpace++

			treeItem = &tree{
				token: &token{
					start:    index + 1,
					end:      treeInitEnd,
					label:    current.lastLabel,
					fragment: quoteFragment,
					symbol:   SymbolValue,
					sub:      tokens{},
				},
				lastLabel: current.lastLabel,
				fragment:  quoteFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level,
			}

			current.lastStart = treeItem.start
			nextStart = index

		default:
			if current.lastStart == treeItem.start {
				nextStart = index
				lastChar = charactor

				continue
			}

			treeItem.end = index

			current.append(treeItem.token)
			nodeSpace++

			treeItem = &tree{
				token: &token{
					start:    index,
					end:      treeInitEnd,
					label:    current.lastLabel,
					fragment: valueFragment,
					symbol:   SymbolValue,
					sub:      tokens{},
				},
				lastLabel: current.lastLabel,
				fragment:  valueFragment,
				parent:    current,
				lastStart: 0,
				level:     current.level,
			}

			current.lastStart = treeItem.start
			nextStart = index
		}
	}

	treeItem.end = root.end

	current.append(treeItem.token)
	nodeSpace++

	// Check the token tree see if there is any tag doesn't closed
	checkLevel := make(tokens, 0, nodeSpace)

	checkLevel = append(checkLevel, root)

	for {
		if len(checkLevel) <= 0 {
			break
		}

		curToken := checkLevel[0]

		if curToken.end == treeInitEnd {
			return newSyntaxErr(
				ErrTreeTokenNotClosed, curToken.start-1, p.input)
		}

		checkLevel = append(checkLevel[1:], curToken.sub...)
	}

	p.root = root
	p.nodeSpace = nodeSpace

	return nil
}

// Value returns parsed parameter values
func (p *parameter) Value() *Value {
	return newValue(p.root, p.input)
}

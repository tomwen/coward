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

package common

import "sync"

// Averager is a dynamically updating Average counter
type Averager interface {
	Get() float64
	Add(value float64) float64
	AddWithWeight(
		value float64, weigher func(average float64, size int) int) float64
}

// averager implements Averager
type averager struct {
	average float64
	total   float64
	cap     int
	size    int
	cur     int
	record  []float64
}

// lockedAverage is same as averager, but it will apply a lock
// when updating and reading
type lockedAverage struct {
	averager

	rwLock sync.RWMutex
}

// NewAverager creates a new Averager
func NewAverager(cap int) Averager {
	return &averager{
		average: 0.0,
		total:   0.0,
		cap:     cap,
		size:    0,
		cur:     0,
		record:  make([]float64, cap),
	}
}

// NewLockedAverager creates a new Averager which applys a
// lock when updating and reading
func NewLockedAverager(cap int) Averager {
	return &lockedAverage{
		averager: averager{
			average: 0.0,
			total:   0.0,
			cap:     cap,
			size:    0,
			cur:     0,
			record:  make([]float64, cap),
		},
		rwLock: sync.RWMutex{},
	}
}

// bump update meta data for next round
func (a *averager) bump() {
	if a.size < a.cap {
		a.size++
	}

	a.cur++

	if a.cur < a.cap {
		return
	}

	a.cur = 0
}

// Get gets current average number
func (a *averager) Get() float64 {
	return a.average
}

// Add add a new number to average counter
func (a *averager) Add(value float64) float64 {
	a.total -= a.record[a.cur]

	a.record[a.cur] = value

	a.total += value

	a.bump()

	a.average = a.total / float64(a.size)

	return a.average
}

// AddWithWeight add the new value to the averager and give
// some weight to the value (Make it count more than just a
// single simple)
func (a *averager) AddWithWeight(
	value float64, weigher func(average float64, size int) int) float64 {
	for times := weigher(a.average, a.size); times > 0; times-- {
		a.Add(value)
	}

	return a.average
}

// Get gets current average number
func (a *lockedAverage) Get() float64 {
	a.rwLock.RLock()

	defer a.rwLock.RUnlock()

	return a.averager.Get()
}

// Add add a new number to average counter
func (a *lockedAverage) Add(value float64) float64 {
	a.rwLock.Lock()

	defer a.rwLock.Unlock()

	return a.averager.Add(value)
}

// AddWithWeight add the new value to the averager and give
// some weight to the value (Make it count more than just a
// single simple)
func (a *lockedAverage) AddWithWeight(
	value float64, weigher func(average float64, size int) int) float64 {
	a.rwLock.Lock()

	defer a.rwLock.Unlock()

	return a.averager.AddWithWeight(value, weigher)
}

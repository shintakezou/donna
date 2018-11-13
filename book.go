// Copyright (c) 2014-2018 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.
//
// I am making my contributions/submissions to this project solely in my
// personal capacity and am not conveying any rights to any intellectual
// property of any third parties.

package donna

import (
	`encoding/binary`
	`math/rand`
	`os`
	`sort`
	`time`
)

// Many pages make a thick book.
type Book struct {
	fileName string
	entries  int64
}

// Opening book record: the fields are exported for binary.Read().
type Entry struct {
	Key   uint64
	Move  uint16
	Score uint16
	Learn uint32
}

func NewBook(bookFile string) (*Book, error) {
	book := &Book{fileName: bookFile}

	if fi, err := os.Stat(book.fileName); err != nil {
		return nil, err
	} else {
		book.entries = fi.Size() / 16
	}

	return book, nil
}

func (b *Book) pickMove(position *Position) Move {
	// The generation of random numbers is too important to be left to chance.
	// Returns pseudo-random integer in [0, limit] range. It panics if limit <= 0.
	random := func(limit int) int {
		rand.Seed(time.Now().Unix()); return rand.Intn(limit)
	}

	entries := b.lookup(position)
	switch length := len(entries); length {
	case 0:
		// TODO: set the "useless book" flag after a few misses.
		return Move(0)
	case 1:
		// The only move available.
		return b.move(position, entries[0])
	default:
		// Sort book entries by score and pick among two best moves.
		sort.Sort(byBookScore{entries})
		best := min(2, len(entries))
		return b.move(position, entries[random(best)])
	}
}

func (b *Book) lookup(position *Position) (entries []Entry) {
	var entry Entry

	file, err := os.Open(b.fileName)
	if err != nil {
		return entries
	}
	defer file.Close()

	key := position.id

	// Since book entries are ordered by polyglot key we can use binary
	// search to find *first* book entry that matches the position.
	first, current, last := int64(-1), int64(0), b.entries
	for first < last {
		current = (first + last) / 2
		file.Seek(current * 16, 0)
		binary.Read(file, binary.BigEndian, &entry)
		if key <= entry.Key {
			last = current
		} else {
			first = current + 1
		}
	}

	// Read all book entries for the given position.
	file.Seek(first * 16, 0)
	for {
		binary.Read(file, binary.BigEndian, &entry)
		if key != entry.Key {
			break
		} else {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (b *Book) move(p *Position, entry Entry) Move {
	from, to := entry.from(), entry.to()

	// Check if this is a castle move. In Polyglot they are represented
	// as E1-H1, E1-A1, E8-H8, and E8-A8.
	if from == E1 && to == H1 {
		return NewCastle(p, from, G1)
	} else if from == E1 && to == A1 {
		return NewCastle(p, from, C1)
	} else if from == E8 && to == H8 {
		return NewCastle(p, from, G8)
	} else if from == E8 && to == A8 {
		return NewCastle(p, from, C8)
	}

	move := NewMove(p, from, to)
	if promo := entry.promoted(); promo != 0 {
		move.promote(promo)
	}

	return move
}

// Converts polyglot encoded "from" coordinate to our square.
func (e *Entry) from() int {
	return square(int((e.Move >> 9) & 7), int((e.Move >> 6) & 7))
}

// Converts polyglot encoded "to" coordinate to our square.
func (e *Entry) to() int {
	return square(int((e.Move >> 3) & 7), int(e.Move & 7))
}

// Polyglot encodes "promotion piece" as follows:
//   knight  1 => 4
//   bishop  2 => 6
//   rook    3 => 8
//   queen   4 => 10
func (e *Entry) promoted() int {
	piece := int((e.Move >> 12) & 7)
	if piece == 0 {
		return piece
	}
	return piece * 2 + 2
}

type byBookScore struct {
	list []Entry
}

func (a byBookScore) Len() int           { return len(a.list) }
func (a byBookScore) Swap(i, j int)      { a.list[i], a.list[j] = a.list[j], a.list[i] }
func (a byBookScore) Less(i, j int) bool { return a.list[i].Score > a.list[j].Score }

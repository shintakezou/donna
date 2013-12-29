// Copyright (c) 2013 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import (
        `fmt`
        `regexp`
)

type Move struct {
        From     int
        To       int
        Piece    Piece
        Captured Piece
        Promoted Piece
}

func NewMove(from, to int, moved, captured Piece) *Move {
        move := new(Move)

        move.From = from
        move.To = to
        move.Piece = moved
        move.Captured = captured

        return move
}

func NewMoveFromString(e2e4 string, p *Position) (move *Move) {
	re := regexp.MustCompile(`([KQRBN]?)([a-h])([1-8])-?([a-h])([1-8])[QRBN]?`)
	arr := re.FindStringSubmatch(e2e4)

	if len(arr) > 0 {
		piece   := arr[1]
		from    := Square(int(arr[3][0]-'1'), int(arr[2][0]-'a'))
                to      := Square(int(arr[5][0]-'1'), int(arr[4][0]-'a'))
		capture := p.pieces[to]

		switch piece {
		case `K`:
			move = NewMove(from, to, King(p.color), capture)
		case `Q`:
			move = NewMove(from, to, Queen(p.color), capture)
		case `R`:
			move = NewMove(from, to, Rook(p.color), capture)
		case `B`:
			move = NewMove(from, to, Bishop(p.color), capture)
		case `N`:
			move = NewMove(from, to, Knight(p.color), capture)
		default:
			move = NewMove(from, to, Pawn(p.color), capture)
		}
                if (p.pieces[from] != move.Piece) || (p.targets[from] & Shift(to) == 0) {
                        move = nil
                }
	} else if e2e4 == `0-0` || e2e4 == `0-0-0` {
                from := p.outposts[King(p.color)].FirstSet()
                to := G1
                if e2e4 == `0-0-0` {
                        to = C1
                }
                if p.color == BLACK {
                        to += 56
                }
                move = NewMove(from, to, King(p.color), 0)
                if !move.isCastle() {
                        move = nil
                }
	}
	return
}

func (m *Move) score(position *Position) int {
	var midgame, endgame int
	square := flip[m.Piece.Color()][m.To]

	switch m.Piece.Kind() {
        case PAWN:
                midgame += bonusPawn[0][square]
                endgame += bonusPawn[1][square]
        case KNIGHT:
                midgame += bonusKnight[0][square]
                endgame += bonusKnight[1][square]
        case BISHOP:
                midgame += bonusBishop[0][square]
                endgame += bonusBishop[1][square]
        case ROOK:
                midgame += bonusRook[0][square]
                endgame += bonusRook[1][square]
        case QUEEN:
                midgame += bonusQueen[0][square]
                endgame += bonusQueen[1][square]
        case KING:
                midgame += bonusKing[0][square]
                endgame += bonusKing[1][square]
        }

	return (midgame * position.stage + endgame * (256 - position.stage)) / 256
}

// PxQ, NxQ, BxQ, RxQ, QxQ, KxQ => where => QUEEN  = 5 << 1 // 10
// PxR, NxR, BxR, RxR, QxR, KxR             ROOK   = 4 << 1 // 8
// PxB, NxB, BxB, RxB, QxB, KxB             BISHOP = 3 << 1 // 6
// PxN, NxN, BxN, RxN, QxN, KxN             KNIGHT = 2 << 1 // 4
// PxP, NxP, BxP, RxP, QxP, KxP             PAWN   = 1 << 1 // 2
func (m *Move) value() int {
        if m.Captured == 0 || m.Captured.Kind() == KING {
                return 0
        }

        victim := (QUEEN - m.Captured.Kind()) / PAWN
        attacker := m.Piece.Kind() / PAWN - 1

        return victimAttacker[victim][attacker]
}

func (m *Move) Promote(kind int) *Move {
        m.Promoted = Piece(kind | m.Piece.Color())

        return m
}

func (m *Move) isValid(p *Position) bool {
        if (m.isKingSideCastle() && !p.isKingSideCastleAllowed()) || (m.isQueenSideCastle() && !p.isQueenSideCastleAllowed()) {
                return false
        }
        return true
}

func (m *Move) isKingSideCastle() bool {
        return m.Piece.IsKing() && ((m.Piece.IsWhite() && m.From == E1 && m.To == G1) || (m.Piece.IsBlack() && m.From == E8 && m.To == G8))
}

func (m *Move) isQueenSideCastle() bool {
        return m.Piece.IsKing() && ((m.Piece.IsWhite() && m.From == E1 && m.To == C1) || (m.Piece.IsBlack() && m.From == E8 && m.To == C8))
}

func (m *Move) isCastle() bool {
        return m.isKingSideCastle() || m.isQueenSideCastle()
}

func (m *Move) isEnpassant(opponentPawns Bitmask) bool {
        color := m.Piece.Color()

        if m.Piece.IsPawn() && Row(m.From) == [2]int{1,6}[color] && Row(m.To) == [2]int{3,4}[color] {
                switch col := Col(m.To); col {
                case 0:
                        return opponentPawns.IsSet(m.To + 1)
                case 7:
                        return opponentPawns.IsSet(m.To - 1)
                default:
                        return opponentPawns.IsSet(m.To + 1) || opponentPawns.IsSet(m.To - 1)
                }
        }
        return false
}

func (m *Move) isEnpassantCapture(enpassant int) bool {
        return m.Piece.IsPawn() && m.To == enpassant
}

func (m *Move) String() string {

        if !m.isCastle() {
                col := [2]int{ Col(m.From) + 'a', Col(m.To) + 'a' }
                row := [2]int{ Row(m.From) + 1, Row(m.To) + 1 }

                capture := '-'
                if m.Captured != 0 {
                        capture = 'x'
                }
                piece, promoted := m.Piece.String(), m.Promoted.String()
                format := `%c%d%c%c%d%s`

                if m.Piece.IsPawn() { // Skip piece name if it's a pawn.
                        return fmt.Sprintf(format, col[0], row[0], capture, col[1], row[1], promoted)
                } else {
                        if Settings.Fancy { // Fancy notation is more readable with extra space.
                                format = `%s ` + format
                        } else {
                                format = `%s` + format
                        }
                        return fmt.Sprintf(format, piece, col[0], row[0], capture, col[1], row[1], promoted)
                }
        } else if m.isKingSideCastle() {
                return `0-0`
        } else {
                return `0-0-0`
        }
}
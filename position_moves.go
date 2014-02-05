// Copyright (c) 2013 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna
import(`regexp`)

func (p *Position) NewMove(from, to int) Move {
        piece, capture := p.pieces[from], p.pieces[to]

        if p.flags.enpassant != 0 && to == p.flags.enpassant {
                capture = Pawn(piece.color()^1)
        }

        return Move(from | (to << 8) | (int(piece) << 16) | (int(capture) << 20))
}

func (p *Position) NewCastle(from, to int) Move {
        return Move(from | (to << 8) | (int(p.pieces[from]) << 16) | isCastle)
}

func (p *Position) NewEnpassant(from, to int) Move {
        return Move(from | (to << 8) | (int(p.pieces[from]) << 16) | isEnpassant)
}

func (p *Position) NewPawnJump(from, to int) Move {
        return Move(from | (to << 8) | (int(p.pieces[from]) << 16) | isPawnJump)
}

func (p *Position) moveScore(m Move) int {
	square := flip[m.color()][m.to()]
        midgame, endgame := m.piece().bonus(square)

	return (midgame * p.stage + endgame * (256 - p.stage)) / 256
}

func (p *Position) pawnMove(square, target int) Move {
        if RelRow(square, p.color) == 1 && RelRow(target, p.color) == 3 {
                if p.isEnpassant(target, p.color) {
                        return NewEnpassant(p, square, target)
                } else {
                        return NewPawnJump(p, square, target)
                }
        }

        return NewMove(p, square, target)
}

func (p *Position) pawnPromotion(square, target int) (Move, Move, Move, Move) {
        return NewMove(p, square, target).promote(QUEEN),
               NewMove(p, square, target).promote(ROOK),
               NewMove(p, square, target).promote(BISHOP),
               NewMove(p, square, target).promote(KNIGHT)
}

func (p *Position) isEnpassant(target, color int) bool {
        pawns := p.outposts[Pawn(color^1)] // Opposite color pawns.
        switch col := Col(target); col {
        case 0:
                return pawns.isSet(target + 1)
        case 7:
                return pawns.isSet(target - 1)
        default:
                return pawns.isSet(target + 1) || pawns.isSet(target - 1)
        }
        return false
}

func (p *Position) pawnMovesMask(color int) (mask Bitmask) {
        if color == White {
                mask = (p.outposts[Pawn(White)] << 8)
        } else {
                mask = (p.outposts[Pawn(Black)] >> 8)
        }
        mask &= ^p.board[2]
        return
}

func (p *Position) pawnJumpsMask(color int) (mask Bitmask) {
        if color == White {
                mask = maskRank[3] & (p.outposts[Pawn(White)] << 16)
        } else {
                mask = maskRank[4] & (p.outposts[Pawn(Black)] >> 16)
        }
        mask &= ^p.board[2]
        return
}

func (p *Position) NewMoveFromString(e2e4 string) (move Move) {
	re := regexp.MustCompile(`([KkQqRrBbNn]?)([a-h])([1-8])-?([a-h])([1-8])([QqRrBbNn]?)`)
	arr := re.FindStringSubmatch(e2e4)

	if len(arr) > 0 {
		name  := arr[1]
		from  := Square(int(arr[3][0]-'1'), int(arr[2][0]-'a'))
		to    := Square(int(arr[5][0]-'1'), int(arr[4][0]-'a'))
		promo := arr[6]

		var piece Piece
		switch name {
		case `K`, `k`:
			piece = King(p.color)
		case `Q`, `q`:
			piece = Queen(p.color)
		case `R`, `r`:
			piece = Rook(p.color)
		case `B`, `b`:
			piece = Bishop(p.color)
		case `N`, `n`:
			piece = Knight(p.color)
		default:
			piece = p.pieces[from] // <-- Makes piece character optional.
		}
                if (p.pieces[from] != piece) || (p.targets[from] & Bit(to) == 0) {
                        move = 0 // Invalid move.
                } else {
                        move = NewMove(p, from, to)
                        if len(promo) > 0 {
                                switch promo {
                                case `Q`, `q`:
                                        move.promote(QUEEN)
                                case `R`, `r`:
                                        move.promote(ROOK)
                                case `B`, `b`:
                                        move.promote(BISHOP)
                                case `N`, `n`:
                                        move.promote(KNIGHT)
                                default:
                                        move = 0
                                }
                        }
                }
	} else if e2e4 == `0-0` || e2e4 == `0-0-0` {
                from := p.outposts[King(p.color)].first()
                to := G1
                if e2e4 == `0-0-0` {
                        to = C1
                }
                if p.color == Black {
                        to += 56
                }
                move = NewMove(p, from, to)
                if !move.isCastle() {
                        move = 0
                }
	}
	return
}

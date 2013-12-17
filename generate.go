// Copyright (c) 2013 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import()

// All moves.
func (p *Position) Moves() (moves []*Move) {
        for i, piece := range p.pieces {
                if piece != 0 && piece.Color() == p.color {
                        moves = append(moves, p.PossibleMoves(i, piece)...)
                }
        }
        if len(moves) > 1 {
                moves = p.Reorder(moves)
        }
        Log("%d candidates for %s: %v\n", len(moves), C(p.color), moves)

        return
}

func (p *Position) Captures() (moves []*Move) {
        for i, piece := range p.pieces {
                if piece != 0 && piece.Color() == p.color {
                        moves = append(moves, p.PossibleCaptures(i, piece)...)
                }
        }
        Log("%d capture candidates for %s: %v\n", len(moves), C(p.color), moves)

        return
}

// All moves for the piece in certain square.
func (p *Position) PossibleMoves(square int, piece Piece) (moves []*Move) {
        targets := p.targets[square]

        for targets.IsNotEmpty() {
                target := targets.FirstSet()
                capture := p.pieces[target]
                //
                // For regular moves each target square represents one possible
                // move. For pawn promotion, however, we have to generate four
                // possible moves, one for each promoted piece.
                //
                if !p.isPawnPromotion(piece, target) {
                        candidate := NewMove(square, target, piece, capture)
                        if !p.MakeMove(candidate).isCheck(p.color) {
                                moves = append(moves, candidate)
                        }
                } else {
                        for _,name := range([]int{ QUEEN, ROOK, BISHOP, KNIGHT }) {
                                candidate := NewMove(square, target, piece, capture).Promote(name)
                                if !p.MakeMove(candidate).isCheck(p.color) {
                                        moves = append(moves, candidate)
                                }
                        }
                }
                targets.Clear(target)
        }
        if castle := p.tryCastle(); castle != nil {
                moves = append(moves, castle)
        }
        return
}

// All capture moves for the piece in certain square.
func (p *Position) PossibleCaptures(square int, piece Piece) (moves []*Move) {
        targets := p.targets[square]

        for targets.IsNotEmpty() {
                target := targets.FirstSet()
                capture := p.pieces[target]
                if capture != 0  {
                        if !p.isPawnPromotion(piece, target) {
                                candidate := NewMove(square, target, piece, capture)
                                if !p.MakeMove(candidate).isCheck(p.color) {
                                        moves = append(moves, candidate)
                                }
                        } else {
                                for _,name := range([]int{ QUEEN, ROOK, BISHOP, KNIGHT }) {
                                        candidate := NewMove(square, target, piece, capture).Promote(name)
                                        if !p.MakeMove(candidate).isCheck(p.color) {
                                                moves = append(moves, candidate)
                                        }
                                }
                        }
                }
                targets.Clear(target)
        }
        return
}

func (p *Position) Reorder(moves []*Move) []*Move {
        var checks, promotions, captures, remaining []*Move

        for _, move := range moves {
                if p.MakeMove(move).check {
                        checks = append(checks, move)
                } else if move.Promoted != 0 {
                        promotions = append(promotions, move)
                } else if move.Captured != 0 {
                        captures = append(captures, move)
                } else {
                        remaining = append(remaining, move)
                }
        }

        return append(append(append(captures, promotions...), checks...), remaining...)
}

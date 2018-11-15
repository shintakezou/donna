// Copyright (c) 2014-2018 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.
//
// I am making my contributions/submissions to this project solely in my
// personal capacity and am not conveying any rights to any intellectual
// property of any third parties.

package donna

func (gen *MoveGen) generateCaptures() *MoveGen {
	our, their := gen.p.colors()
	return gen.pawnCaptures(our, their).pieceCaptures(our, their)
}

// Generates all pseudo-legal pawn captures and promotions.
func (gen *MoveGen) pawnCaptures(our, their int) *MoveGen {
	opponent := gen.p.outposts[their&1]

	for pawns := gen.p.outposts[pawn(our)]; pawns.anyʔ(); pawns = pawns.pop() {
		square := pawns.first()

		// For pawns on files 2-6 the moves include captures only,
		// while for pawns on the 7th file the moves include captures
		// as well as queen promotion.
		if square.rank(our) != A7H7 {
			gen.movePawn(square, gen.p.targets(square) & opponent)
		} else {
			for bm := gen.p.targets(square); bm.anyʔ(); bm = bm.pop() {
				target := bm.first()
				mQ, _, _, _ := NewPromotion(gen.p, square, target)
				gen.add(mQ)
			}
		}
	}

	return gen
}

// Generates all pseudo-legal captures by pieces other than pawn.
func (gen *MoveGen) pieceCaptures(our, their int) *MoveGen {
	opponent := gen.p.outposts[their&1]

	for bm := gen.p.outposts[our&1] ^ gen.p.outposts[pawn(our)] ^ gen.p.outposts[king(our)]; bm.anyʔ(); bm = bm.pop() {
		square := bm.first()
		gen.movePiece(square, gen.p.targets(square) & opponent)
	}
	if gen.p.outposts[king(our)].anyʔ() {
		square := gen.p.king[our&1]
		gen.moveKing(square, gen.p.targets(square) & opponent)
	}

	return gen
}

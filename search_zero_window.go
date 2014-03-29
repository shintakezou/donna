// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import()

var razoringMargin = [4]int{ 0, 240, 450, 660 }
var futilityMargin = [4]int{ 0, 400, 500, 600 }

// Search with zero window.
func (p *Position) searchWithZeroWindow(beta, depth int) int {
        p.game.nodes++
        if p.isRepetition() {
                return 0
        }

        bestScore := Ply() - Checkmate
        if bestScore >= beta {
                return beta//bestScore
        }

        score := p.Evaluate()

        // Razoring and futility pruning. TODO: disable or tune-up in puzzle solving mode.
        if depth < len(razoringMargin) {
                if margin := beta - razoringMargin[depth]; score < margin && beta < 31000 {
                        if p.outposts[pawn(p.color)] & mask7th[p.color] == 0 { // No pawns on 7th.
                                razorScore := p.searchQuiescence(margin - 1, margin)
                                if razorScore < margin {
                                        return razorScore
                                }
                        }
                }

                if margin := score - futilityMargin[depth]; margin >= beta && beta > -31000 {
                        if p.outposts[p.color] & ^p.outposts[king(p.color)] & ^p.outposts[pawn(p.color)] != 0 {
                                return margin
                        }
                }
        }

        // Null move pruning.
        if depth > 1 && score >= beta && p.outposts[p.color].count() > 5 /*&& beta > -31000*/ {
                reduction := 3 + depth / 4
                if score - 100 > beta {
                        reduction++
                }

                position := p.MakeNullMove()
                if depth <= reduction {
                        score = -position.searchQuiescence(-beta, 1 - beta)
                } else {
                        score = -position.searchWithZeroWindow(1 - beta, depth - reduction)
                }
                position.TakeBackNullMove()

                if score >= beta {
                        return score
                }
        }

        moveCount := 0
        gen := p.StartMoveGen(Ply()).GenerateMoves().rank()
        for move := gen.NextMove(); move != 0; move = gen.NextMove() {
                if position := p.MakeMove(move); position != nil {
                        //Log("%*szero/%s> depth: %d, ply: %d, move: %s\n", Ply()*2, ` `, C(p.color), depth, Ply(), move)
                        inCheck, giveCheck := position.isInCheck(position.color), position.isInCheck(position.color^1)

                        reducedDepth := depth
                        if !inCheck && !giveCheck && move.capture() == 0 && move.promo() == 0 && depth >= 3 && moveCount >= 8 {
                                reducedDepth = depth - 2 // Late move reduction. TODO: disable or tune-up in puzzle solving mode.
                                if reducedDepth > 0 && moveCount >= 16 {
                                        reducedDepth--
                                        if reducedDepth > 0 && moveCount >= 32 {
                                                reducedDepth--
                                        }
                                }
                        } else if !inCheck {
                                reducedDepth = depth - 1
                        }

                        moveScore := 0
                        if reducedDepth == 0 {
                                moveScore = -position.searchQuiescence(-beta, 1 - beta)
                        } else if inCheck {
                                moveScore = -position.searchInCheck(1 - beta, reducedDepth)
                        } else {
                                moveScore = -position.searchWithZeroWindow(1 - beta, reducedDepth)

                                // Verify late move reduction.
                                if reducedDepth < depth - 1 && moveScore >= beta {
                                        moveScore = -position.searchWithZeroWindow(1 - beta, depth - 1)
                                }
                        }

                        position.TakeBack(move)
                        moveCount++

                        if moveScore > bestScore {
                                if moveScore >= beta {
                                        if move.capture() == 0 && move.promo() == 0 && move != p.killers[0] {
                                                p.killers[1] = p.killers[0]
                                                p.killers[0] = move
                                        	p.game.goodMoves[move.piece()][move.to()] += depth * depth;
                                                //Log(">>> depth: %d, node: %d, killers %s/%s\n", depth, node, p.killers[0], p.killers[1])
                                        }
                                        return moveScore
                                }
                                bestScore = moveScore
                        }
                }
        } // next move.

        if moveCount == 0 {
                return 0
        }

        return bestScore
}

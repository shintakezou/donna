// Copyright (c) 2014-2018 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.
//
// I am making my contributions/submissions to this project solely in my
// personal capacity and am not conveying any rights to any intellectual
// property of any third parties.

package donna

import (
	`fmt`
	`time`
)

// Poor man's ternary. Works best with scalar `yes` and `no`.
func let(ok bool, yes, no int) int {
	if ok {
		return yes
	}

	return no
}

// Returns a distance between current node and the root one.
func ply() int {
	return node - rootNode
}

// Returns a score of getting mated in given number of plies.
func matedIn(ply int) int {
	return ply - Checkmate
}

// Returns a score of mating an opponent in given number of plies.
func matingIn(ply int) int {
	return Checkmate - ply
}

// Adjusts values of alpha and beta based on how close we are
// to checkmate or be checkmated.
func mateDistance(alpha, beta, ply int) (int, int) {
	return max(matedIn(ply), alpha), min(matingIn(ply + 1), beta)
}

func mateʔ(score int) bool {
	return abs(score) >= Checkmate - MaxPly
}

// Integer version of math/abs.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func max64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// Returns time in milliseconds elapsed since the given start time.
func since(start time.Time) int64 {
	return time.Since(start).Nanoseconds() / 1000000
}

// Returns nodes per second search speed for the given time duration.
func nps(duration int64) int64 {
	nodes := int64(game.nodes + game.qnodes) * 1000
	if duration != 0 {
		return nodes / duration
	}
	return nodes
}

// Returns a number of used items in a sample of 1000 cache entries.
func hashfull() int {
	count := 0

	if cacheSize := len(game.cache); cacheSize > 1000 {
		start := (game.nodes + game.qnodes) % (cacheSize - 1000) // 0 <= start < cacheSize - 1000.
		for i := start; i < start + 1000; i++ {
			if game.cache[i].token() == game.token {
				count++
			}
		}
	}

	return count
}

// Formats time duration in milliseconds in human readable form (MM:SS.XXX).
func ms(duration int64) string {
	mm := duration / 1000 / 60
	ss := duration / 1000 % 60
	xx := duration - mm * 1000 * 60 - ss * 1000

	return fmt.Sprintf(`%02d:%02d.%03d`, mm, ss, xx)
}

func C(color int) string {
	if color == White {
		return `white`
	}
	return `black`
}

func Summary(metrics map[string]interface{}) {
	phase := metrics[`Phase`].(int)
	tally := metrics[`PST`].(Score)
	material := metrics[`Imbalance`].(Score)
	final := metrics[`Final`].(Score)
	units := float32(onePawn)

	fmt.Println()
	fmt.Printf("Metric              MidGame        |        EndGame        | Blended\n")
	fmt.Printf("                W      B     W-B   |    W      B     W-B   |  (%d)  \n", phase)
	fmt.Printf("-----------------------------------+-----------------------+--------\n")
	fmt.Printf("%-12s    -      -    %5.2f  |    -      -    %5.2f  >  %5.2f\n", `PST`,
		float32(tally.midgame)/units, float32(tally.endgame)/units, float32(tally.blended(phase))/units)
	fmt.Printf("%-12s    -      -    %5.2f  |    -      -    %5.2f  >  %5.2f\n", `Imbalance`,
		float32(material.midgame)/units, float32(material.endgame)/units, float32(material.blended(phase))/units)

	for _, tag := range([]string{`Tempo`, `Center`, `Threats`, `Pawns`, `Passers`, `Mobility`, `+Pieces`, `-Knights`, `-Bishops`, `-Rooks`, `-Queens`, `+King`, `-Cover`, `-Safety`}) {
		white := metrics[tag].(Total).white
		black := metrics[tag].(Total).black

		var score Score
		score.add(white).sub(black)

		if tag[0:1] == `+` {
			tag = tag[1:]
		} else if tag[0:1] == `-` {
			tag = `  ` + tag[1:]
		}

		fmt.Printf("%-12s  %5.2f  %5.2f  %5.2f  |  %5.2f  %5.2f  %5.2f  >  %5.2f\n", tag,
			float32(white.midgame)/units, float32(black.midgame)/units, float32(score.midgame)/units,
			float32(white.endgame)/units, float32(black.endgame)/units, float32(score.endgame)/units,
			float32(score.blended(phase))/units)
	}
	fmt.Printf("%-12s    -      -    %5.2f  |    -      -    %5.2f  >  %5.2f\n\n", `Final Score`,
		float32(final.midgame)/units, float32(final.endgame)/units, float32(final.blended(phase))/units)
}

// Logging wrapper around `fmt.Printf()` that could be turned on as needed. Typical
// usage is `log(); defer log()` in tests.
func log(args ...interface{}) {
	switch len(args) {
	case 0:
		// Calling Log() with no arguments flips the logging setting.
		engine.logʔ = !engine.logʔ
		engine.fancyʔ = !engine.fancyʔ
	case 1:
		switch args[0].(type) {
		case bool:
			engine.logʔ = args[0].(bool)
			engine.fancyʔ = args[0].(bool)
		default:
			if engine.logʔ {
				fmt.Println(args[0])
			}
		}
	default:
		if engine.logʔ {
			fmt.Printf(args[0].(string), args[1:]...)
		}
	}
}

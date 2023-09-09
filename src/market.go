package main

import (
	"errors"
	"math"
)

// logarithmic market scoring rule (LMSR) market maker from Robin Hanson:
// https://mason.gmu.edu/~rhanson/mktscore.pdf
func BinaryLMSRBuy(invariant int, funding int, tokensAQ float64, tokensBQ float64, sats int) (float64, error) {
	k := float64(invariant)
	f := float64(funding)
	numOutcomes := 2.0
	expA := -tokensAQ / f
	expB := -tokensBQ / f
	if k != math.Pow(numOutcomes, expA)+math.Pow(numOutcomes, expB) {
		// invariant should not already be broken
		return -1, errors.New("invariant already broken")
	}
	// AMM converts order into equal amount of tokens per outcome and then solves equation to fix invariant
	// see https://docs.gnosis.io/conditionaltokens/docs/introduction3/#an-example-with-lmsr
	newTokensA := tokensAQ + float64(sats)
	newTokensB := tokensBQ + float64(sats)
	expB = -newTokensB / f
	x := newTokensA + f*math.Log(k-math.Pow(numOutcomes, expB))/math.Log(numOutcomes)
	expA = -(newTokensA - x) / f
	if k != math.Pow(numOutcomes, expA)+math.Pow(numOutcomes, expB) {
		// invariant should not be broken
		return -1, errors.New("invariant broken")
	}
	return x, nil
}
